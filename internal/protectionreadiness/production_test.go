package protectionreadiness

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProductionProviderDefaultsPairedUnwiredWithoutPinnedManifest(t *testing.T) {
	provider := NewProductionProvider(ProductionConfig{ConfigDir: t.TempDir()})
	snapshot, err := provider.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		if verdict := snapshot.Verdict(market); verdict.State != Unwired || verdict.Code != RefusalMissingEvidence {
			t.Fatalf("%s default=%+v", market, verdict)
		}
	}
}

func TestProductionProviderPreservesExactScopeAndCachesAcceptedSerial(t *testing.T) {
	fixture := newProductionFixture(t)
	provider := NewProductionProvider(fixture.config)
	if contracts := provider.RuntimeContracts(); len(contracts) != 2 || contracts[0].Market != MarketKR || contracts[1].Market != MarketUS {
		t.Fatalf("paired contracts=%+v", contracts)
	}
	for check := 0; check < 2; check++ {
		snapshot, err := provider.Current(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for _, market := range []Market{MarketKR, MarketUS} {
			if verdict := snapshot.Verdict(market); verdict.State != Wired || verdict.Code != RefusalNone {
				t.Fatalf("check %d %s=%+v", check, market, verdict)
			}
			scope := exactDispatchScope(snapshot, market, 50)
			if decision := snapshot.Dispatch(scope, fixture.now); !decision.Allowed {
				t.Fatalf("check %d %s dispatch=%+v", check, market, decision)
			}
		}
	}
}

func TestProductionProviderIsolatesMarketEvidenceDriftButClosesPairedOnGlobalStateCorruption(t *testing.T) {
	fixture := newProductionFixture(t)
	provider := NewProductionProvider(fixture.config)
	first, _ := provider.Current(context.Background())
	if first.Verdict(MarketKR).State != Wired || first.Verdict(MarketUS).State != Wired {
		t.Fatalf("initial=%+v %+v", first.Verdict(MarketKR), first.Verdict(MarketUS))
	}
	if err := os.WriteFile(filepath.Join(fixture.dir, "kr-evidence.json"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	second, _ := provider.Current(context.Background())
	if got := second.Verdict(MarketKR); got.State != Unwired || got.Code == RefusalNone {
		t.Fatalf("KR drift=%+v", got)
	}
	if got := second.Verdict(MarketUS); got.State != Wired || got.Code != RefusalNone {
		t.Fatalf("KR drift contaminated US=%+v", got)
	}

	if err := os.WriteFile(filepath.Join(fixture.dir, productionStateFile), []byte(`{"schema_version":"corrupt","serials":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	restarted := NewProductionProvider(fixture.config)
	closed, _ := restarted.Current(context.Background())
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := closed.Verdict(market); got.State != Unwired || got.Code != RefusalStateCorrupt {
			t.Fatalf("global state corruption %s=%+v", market, got)
		}
	}
}

type productionFixture struct {
	dir    string
	now    time.Time
	config ProductionConfig
}

func newProductionFixture(t *testing.T) productionFixture {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	now := readinessNow
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	buildDigest, toolDigest := digestOf("production-build"), digestOf("production-tool")
	assemblies := []SupervisorAssembly{
		{Market: MarketKR, ComponentDigest: digestOf("production-supervisor-KR"), Wired: true},
		{Market: MarketUS, ComponentDigest: digestOf("production-supervisor-US"), Wired: true},
	}
	markets := make([]productionMarketConfig, 0, 2)
	for _, market := range []Market{MarketKR, MarketUS} {
		lower := "kr"
		if market == MarketUS {
			lower = "us"
		}
		evidenceName := lower + "-evidence.json"
		evidence := []byte(`{"result":"verified"}`)
		writeProductionFile(t, filepath.Join(dir, evidenceName), evidence, 0o600)
		body := attestationBody{SchemaVersion: SchemaVersionV1, Serial: 1, KeyID: "production-key", SignatureAlgorithm: AlgorithmEd25519,
			AccountID: "acct", ProfileID: "production", Market: market, OrderType: "LIMIT", SessionScope: "REGULAR",
			QuantityMin: 1, QuantityMax: 100, TriggerSource: "LAST_TRADE", ReplaceSemantics: ReplaceAtomic,
			Broker: fixtureBrokerCapability(), ToolDigest: toolDigest, BuildDigest: buildDigest, EvidenceDigest: sha256Hex(evidence),
			IssuedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339Nano)}
		canonical, _ := canonicalAttestationBody(body)
		envelope, _ := json.Marshal(attestationEnvelope{attestationBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))})
		attestationName := lower + "-attestation.json"
		writeProductionFile(t, filepath.Join(dir, attestationName), envelope, 0o600)
		markets = append(markets, productionMarketConfig{Market: market, OrderType: body.OrderType, SessionScope: body.SessionScope,
			QuantityMin: body.QuantityMin, QuantityMax: body.QuantityMax, TriggerSource: body.TriggerSource, ReplaceSemantics: body.ReplaceSemantics,
			Broker: body.Broker, SupervisorDigest: assemblies[len(markets)].ComponentDigest, AttestationFile: attestationName, EvidenceFile: evidenceName})
	}
	manifest := productionManifest{SchemaVersion: productionSchema, MaximumLifetimeSeconds: 7200, MaximumOverlapSeconds: 600,
		Keys: []productionKey{{KeyID: "production-key", PublicKey: base64.StdEncoding.EncodeToString(public),
			AcceptFrom: now.Add(-time.Hour).Format(time.RFC3339Nano), PrimaryUntil: now.Add(2 * time.Hour).Format(time.RFC3339Nano), OverlapUntil: now.Add(2 * time.Hour).Format(time.RFC3339Nano)}},
		Markets: markets}
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeProductionFile(t, filepath.Join(dir, ProductionManifestFile), manifestData, 0o400)
	return productionFixture{dir: dir, now: now, config: ProductionConfig{ConfigDir: dir, AccountID: "acct", ProfileID: "production",
		BuildDigest: buildDigest, ToolDigest: toolDigest, ManifestDigest: sha256Hex(manifestData), Now: func() time.Time { return now }, SupervisorAssemblies: assemblies}}
}

func writeProductionFile(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}
