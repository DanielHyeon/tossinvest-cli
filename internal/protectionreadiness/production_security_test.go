//go:build unix

package protectionreadiness

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProductionCacheReevaluatesAttestationTimeBoundaries(t *testing.T) {
	for _, test := range []struct {
		name   string
		want   RefusalCode
		mutate func(*ProductionProvider, time.Time)
		step   time.Duration
	}{
		{name: "attestation expiry", want: RefusalExpired, step: time.Hour},
		{name: "key revocation", want: RefusalRevokedKey, step: time.Minute, mutate: func(provider *ProductionProvider, now time.Time) {
			provider.policy.keys[0].revokedAt = now.Add(time.Minute)
			provider.policy.seal = pinnedPolicySeal(provider.policy)
		}},
		{name: "rotation overlap end", want: RefusalRotationWindow, step: time.Minute, mutate: func(provider *ProductionProvider, now time.Time) {
			provider.policy.keys[0].primaryUntil = now.Add(30 * time.Second)
			provider.policy.keys[0].overlapUntil = now.Add(time.Minute)
			provider.policy.seal = pinnedPolicySeal(provider.policy)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newProductionFixture(t)
			now := fixture.now
			fixture.config.Now = func() time.Time { return now }
			provider := NewProductionProvider(fixture.config)
			first, err := provider.Current(context.Background())
			if err != nil || first.Verdict(MarketKR).State != Wired {
				t.Fatalf("initial snapshot=%+v err=%v", first.Verdict(MarketKR), err)
			}
			if test.mutate != nil {
				test.mutate(provider, now)
			}
			now = now.Add(test.step)
			second, err := provider.Current(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got := second.Verdict(MarketKR); got.State != Unwired || got.Code != test.want {
				t.Fatalf("cached boundary remained authoritative: %+v, want %s", got, test.want)
			}
		})
	}
}

func TestProductionCacheReevaluatesPeerTimeBoundaryWhenOtherMarketArtifactChanges(t *testing.T) {
	for _, test := range []struct {
		name   string
		want   RefusalCode
		mutate func(*ProductionProvider, time.Time)
		step   time.Duration
	}{
		{name: "attestation expiry", want: RefusalExpired, step: time.Hour},
		{name: "key revocation", want: RefusalRevokedKey, step: time.Minute, mutate: func(provider *ProductionProvider, now time.Time) {
			provider.policy.keys[0].revokedAt = now.Add(time.Minute)
			provider.policy.seal = pinnedPolicySeal(provider.policy)
		}},
		{name: "rotation overlap end", want: RefusalRotationWindow, step: time.Minute, mutate: func(provider *ProductionProvider, now time.Time) {
			provider.policy.keys[0].primaryUntil = now.Add(30 * time.Second)
			provider.policy.keys[0].overlapUntil = now.Add(time.Minute)
			provider.policy.seal = pinnedPolicySeal(provider.policy)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newProductionFixture(t)
			now := fixture.now
			fixture.config.Now = func() time.Time { return now }
			provider := NewProductionProvider(fixture.config)
			first, err := provider.Current(context.Background())
			if err != nil || first.Verdict(MarketUS).State != Wired {
				t.Fatalf("initial US snapshot=%+v err=%v", first.Verdict(MarketUS), err)
			}
			if test.mutate != nil {
				test.mutate(provider, now)
			}
			if err := os.WriteFile(filepath.Join(fixture.dir, "kr-evidence.json"), []byte("changed"), 0o600); err != nil {
				t.Fatal(err)
			}
			now = now.Add(test.step)
			second, err := provider.Current(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got := second.Verdict(MarketUS); got.State != Unwired || got.Code != test.want {
				t.Fatalf("unchanged US cache crossed boundary while KR changed: %+v, want %s", got, test.want)
			}
		})
	}
}

func TestProductionProvidersSerializeMonotonicStateAcrossInstances(t *testing.T) {
	fixture := newProductionFixture(t)
	providers := []*ProductionProvider{NewProductionProvider(fixture.config), NewProductionProvider(fixture.config)}
	start := make(chan struct{})
	type result struct {
		snapshot ReadinessSnapshot
		err      error
	}
	results := make(chan result, len(providers))
	for _, provider := range providers {
		go func(provider *ProductionProvider) {
			<-start
			snapshot, err := provider.Current(context.Background())
			results <- result{snapshot: snapshot, err: err}
		}(provider)
	}
	close(start)
	wired, rollback := 0, 0
	for range providers {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		verdict := result.snapshot.Verdict(MarketKR)
		switch {
		case verdict.State == Wired && verdict.Code == RefusalNone:
			wired++
		case verdict.State == Unwired && verdict.Code == RefusalSerialRollback:
			rollback++
		default:
			t.Fatalf("unexpected cross-instance verdict=%+v", verdict)
		}
	}
	if wired != 1 || rollback != 1 {
		t.Fatalf("wired=%d rollback=%d, want one serialized winner", wired, rollback)
	}
}

func TestCachedProviderRejectsNewerDurableSerialFromPeer(t *testing.T) {
	fixture := newProductionFixture(t)
	provider := NewProductionProvider(fixture.config)
	if got := mustCurrent(t, provider).Verdict(MarketKR); got.State != Wired {
		t.Fatalf("initial=%+v", got)
	}
	statePath := filepath.Join(fixture.dir, productionStateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state productionState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	for index := range state.Serials {
		if state.Serials[index].Market == MarketKR {
			state.Serials[index].Serial++
		}
	}
	data, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	writeProductionFile(t, statePath, data, 0o600)
	if got := mustCurrent(t, provider).Verdict(MarketKR); got.State != Unwired || got.Code != RefusalSerialRollback {
		t.Fatalf("stale cached serial remained authoritative: %+v", got)
	}
}

func TestProductionStateMissingAfterBootstrapIsCorrupt(t *testing.T) {
	fixture := newProductionFixture(t)
	provider := NewProductionProvider(fixture.config)
	first, err := provider.Current(context.Background())
	if err != nil || first.Verdict(MarketKR).State != Wired {
		t.Fatalf("bootstrap snapshot=%+v err=%v", first.Verdict(MarketKR), err)
	}
	if err := os.Remove(filepath.Join(fixture.dir, productionStateFile)); err != nil {
		t.Fatal(err)
	}
	second, err := provider.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := second.Verdict(market); got.State != Unwired || got.Code != RefusalStateCorrupt {
			t.Fatalf("missing bootstrapped state %s=%+v", market, got)
		}
	}
}

func TestProductionProviderRejectsSymlinkedOrGroupWritableParent(t *testing.T) {
	fixture := newProductionFixture(t)
	parent := t.TempDir()
	link := filepath.Join(parent, "linked-config")
	if err := os.Symlink(fixture.dir, link); err != nil {
		t.Fatal(err)
	}
	linked := fixture.config
	linked.ConfigDir = link
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := mustCurrent(t, NewProductionProvider(linked)).Verdict(market); got.State != Unwired || got.Code == RefusalNone {
			t.Fatalf("symlinked parent %s=%+v", market, got)
		}
	}

	if err := os.Chmod(fixture.dir, 0o770); err != nil {
		t.Fatal(err)
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := mustCurrent(t, NewProductionProvider(fixture.config)).Verdict(market); got.State != Unwired || got.Code == RefusalNone {
			t.Fatalf("group-writable parent %s=%+v", market, got)
		}
	}
}

func TestProductionProviderRejectsSymlinkInAncestorPath(t *testing.T) {
	fixture := newProductionFixture(t)
	linkRoot := t.TempDir()
	if err := os.Chmod(linkRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	ancestor := filepath.Join(linkRoot, "ancestor")
	if err := os.Symlink(filepath.Dir(fixture.dir), ancestor); err != nil {
		t.Fatal(err)
	}
	linked := fixture.config
	linked.ConfigDir = filepath.Join(ancestor, filepath.Base(fixture.dir))
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := mustCurrent(t, NewProductionProvider(linked)).Verdict(market); got.State != Unwired || got.Code == RefusalNone {
			t.Fatalf("symlinked ancestor %s=%+v", market, got)
		}
	}
}

func TestInvalidSecondMarketPublishesNoPartialRuntimeContract(t *testing.T) {
	fixture := newProductionFixture(t)
	manifestPath := filepath.Join(fixture.dir, ProductionManifestFile)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest productionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Markets[1].Market = MarketKR
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(manifestPath, 0o600); err != nil {
		t.Fatal(err)
	}
	writeProductionFile(t, manifestPath, data, 0o400)
	fixture.config.ManifestDigest = sha256Hex(data)
	provider := NewProductionProvider(fixture.config)
	if contracts := provider.RuntimeContracts(); len(contracts) != 0 {
		t.Fatalf("invalid paired manifest published partial contracts: %+v", contracts)
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := mustCurrent(t, provider).Verdict(market); got.State != Unwired || got.Code != RefusalInvalid {
			t.Fatalf("invalid paired manifest %s=%+v", market, got)
		}
	}
}

func mustCurrent(t *testing.T, provider *ProductionProvider) ReadinessSnapshot {
	t.Helper()
	snapshot, err := provider.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
