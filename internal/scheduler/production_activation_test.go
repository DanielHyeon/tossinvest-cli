package scheduler

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	testActivationSchema = "strategy-activation-manifest:v1"
	testActivationDomain = "TossOS/strategy-scheduler-activation/ed25519/v1"
)

type testActivationManifestBody struct {
	SchemaVersion      string       `json:"schema_version"`
	Domain             string       `json:"domain"`
	SignatureAlgorithm string       `json:"signature_algorithm"`
	KeyID              string       `json:"key_id"`
	Generation         uint64       `json:"generation"`
	SchedulerVersion   string       `json:"scheduler_version"`
	DesiredRevision    uint64       `json:"desired_revision"`
	CalendarVersion    string       `json:"calendar_version"`
	Market             MarketScope  `json:"market"`
	Session            SessionScope `json:"session"`
	ConfigVersion      string       `json:"config_version"`
	BuildDigest        string       `json:"build_digest"`
	Actor              string       `json:"actor"`
	ApprovedAt         string       `json:"approved_at"`
	IssuedAt           string       `json:"issued_at"`
	ExpiresAt          string       `json:"expires_at"`
	Revoked            bool         `json:"revoked"`
}

type testActivationManifest struct {
	testActivationManifestBody
	Signature string `json:"signature"`
}

type productionActivationFixture struct {
	dir     string
	now     time.Time
	public  ed25519.PublicKey
	private ed25519.PrivateKey
	configs map[MarketScope]ProductionActivationConfig
	desired map[MarketScope]DesiredState
	current map[MarketScope]CurrentBinding
}

func TestProductionActivationRestorePairedKRUSExactSuccess(t *testing.T) {
	fixture := newProductionActivationFixture(t)
	requests := fixture.requests(t)
	result, err := RestorePairedProduction(context.Background(), requests, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatalf("RestorePairedProduction: %v", err)
	}
	for _, market := range []MarketScope{MarketScopeKR, MarketScopeUS} {
		got := result.For(market)
		if !got.Restored || got.Reason != ResumeExactManifest || got.Activation == nil || got.Err != nil {
			t.Fatalf("market=%s result=%+v", market, got)
		}
	}
}

func TestProductionActivationPairedBadKRDoesNotBlockValidUS(t *testing.T) {
	fixture := newProductionActivationFixture(t)
	krPath := filepath.Join(fixture.dir, ProductionActivationFileName(MarketScopeKR))
	data, err := os.ReadFile(krPath)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-2] ^= 1
	if err := os.Chmod(krPath, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(krPath, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(krPath, 0o400); err != nil {
		t.Fatal(err)
	}

	requests := fixture.requests(t)
	result, err := RestorePairedProduction(context.Background(), requests, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatalf("paired restore shape: %v", err)
	}
	if got := result.For(MarketScopeKR); got.Restored || got.Activation != nil || got.Reason != ResumeManifestMismatch {
		t.Fatalf("KR tamper result=%+v", got)
	}
	if got := result.For(MarketScopeUS); !got.Restored || got.Activation == nil || got.Reason != ResumeExactManifest {
		t.Fatalf("US peer result=%+v", got)
	}
}

func TestProductionActivationPairedVerificationStartsBothMarketsConcurrently(t *testing.T) {
	started := make(chan MarketScope, 2)
	releaseKR := make(chan struct{})
	verifiers := map[MarketScope]ActivationVerifier{
		MarketScopeKR: blockingActivationVerifier{market: MarketScopeKR, started: started, release: releaseKR},
		MarketScopeUS: blockingActivationVerifier{market: MarketScopeUS, started: started},
	}
	requests := pairedActivationRequests(verifiers)
	done := make(chan PairedRestoreResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := RestorePairedProduction(context.Background(), requests, func() time.Time { return pairedActivationNow })
		done <- result
		errCh <- err
	}()
	seen := map[MarketScope]bool{}
	for len(seen) < 2 {
		select {
		case market := <-started:
			seen[market] = true
		case <-time.After(time.Second):
			t.Fatalf("both verifiers did not start concurrently: %v", seen)
		}
	}
	select {
	case <-done:
		t.Fatal("paired restore returned before blocked KR verification")
	default:
	}
	close(releaseKR)
	result := <-done
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	for _, market := range []MarketScope{MarketScopeKR, MarketScopeUS} {
		if got := result.For(market); !got.Restored || got.Activation == nil {
			t.Fatalf("market=%s result=%+v", market, got)
		}
	}
}

func TestProductionActivationPairedRejectsInvalidPairBeforeVerifier(t *testing.T) {
	for _, tc := range []struct {
		name    string
		markets [2]MarketScope
	}{
		{name: "duplicate KR", markets: [2]MarketScope{MarketScopeKR, MarketScopeKR}},
		{name: "duplicate US", markets: [2]MarketScope{MarketScopeUS, MarketScopeUS}},
		{name: "missing KR", markets: [2]MarketScope{MarketScopeNone, MarketScopeUS}},
		{name: "missing US", markets: [2]MarketScope{MarketScopeKR, MarketScopeNone}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			verifier := countingActivationVerifier{calls: &calls}
			requests := [2]PairedRestoreRequest{
				activationRequest(tc.markets[0], verifier), activationRequest(tc.markets[1], verifier),
			}
			result, err := RestorePairedProduction(context.Background(), requests, func() time.Time { return pairedActivationNow })
			if !errors.Is(err, ErrPairedActivationInvalid) || calls.Load() != 0 {
				t.Fatalf("err=%v calls=%d", err, calls.Load())
			}
			for _, market := range []MarketScope{MarketScopeKR, MarketScopeUS} {
				got := result.For(market)
				if got.Restored || got.Activation != nil || got.Reason != ResumeVerificationFailed {
					t.Fatalf("market=%s result=%+v", market, got)
				}
			}
		})
	}
}

func TestProductionActivationMissingKRVerifierDoesNotInvalidatePresentUSRequest(t *testing.T) {
	requests := pairedActivationRequests(map[MarketScope]ActivationVerifier{
		MarketScopeUS: blockingActivationVerifier{market: MarketScopeUS, started: make(chan MarketScope, 1)},
	})
	result, err := RestorePairedProduction(context.Background(), requests, func() time.Time { return pairedActivationNow })
	if err != nil {
		t.Fatal(err)
	}
	if got := result.For(MarketScopeKR); got.Restored || got.Activation != nil || got.Reason != ResumeManifestUnavailable {
		t.Fatalf("KR unavailable result=%+v", got)
	}
	if got := result.For(MarketScopeUS); !got.Restored || got.Activation == nil || got.Reason != ResumeExactManifest {
		t.Fatalf("US valid result=%+v", got)
	}
}

func TestProductionActivationManifestRefusalsStayMarketLocal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		market MarketScope
		edit   func(*testActivationManifestBody)
		want   ResumeReason
	}{
		{name: "KR revoked", market: MarketScopeKR, edit: func(body *testActivationManifestBody) { body.Revoked = true }, want: ResumeManifestRevoked},
		{name: "US expired", market: MarketScopeUS, edit: func(body *testActivationManifestBody) { body.ExpiresAt = pairedActivationNow.Format(time.RFC3339Nano) }, want: ResumeManifestExpired},
		{name: "KR wrong market", market: MarketScopeKR, edit: func(body *testActivationManifestBody) { body.Market = MarketScopeUS }, want: ResumeManifestMismatch},
		{name: "US zero generation", market: MarketScopeUS, edit: func(body *testActivationManifestBody) { body.Generation = 0 }, want: ResumeManifestMismatch},
		{name: "KR overlong lifetime", market: MarketScopeKR, edit: func(body *testActivationManifestBody) {
			body.ExpiresAt = pairedActivationNow.Add(25 * time.Hour).Format(time.RFC3339Nano)
		}, want: ResumeManifestMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newProductionActivationFixture(t)
			fixture.rewrite(t, tc.market, tc.edit)
			result, err := RestorePairedProduction(context.Background(), fixture.requests(t), func() time.Time { return fixture.now })
			if err != nil {
				t.Fatal(err)
			}
			if got := result.For(tc.market); got.Restored || got.Activation != nil || got.Reason != tc.want {
				t.Fatalf("failed market result=%+v", got)
			}
			peer := MarketScopeUS
			if tc.market == MarketScopeUS {
				peer = MarketScopeKR
			}
			if got := result.For(peer); !got.Restored || got.Activation == nil || got.Reason != ResumeExactManifest {
				t.Fatalf("peer market result=%+v", got)
			}
		})
	}
}

func TestProductionActivationManifestRequiresOwnerOnlyRegularFileAndExactDigest(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*productionActivationFixture, MarketScope)
	}{
		{name: "mode", edit: func(f *productionActivationFixture, market MarketScope) {
			if err := os.Chmod(filepath.Join(f.dir, ProductionActivationFileName(market)), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "digest", edit: func(f *productionActivationFixture, market MarketScope) {
			config := f.configs[market]
			config.ManifestDigest = "sha256:" + strings.Repeat("0", 64)
			f.configs[market] = config
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newProductionActivationFixture(t)
			tc.edit(fixture, MarketScopeKR)
			result, err := RestorePairedProduction(context.Background(), fixture.requests(t), func() time.Time { return fixture.now })
			if err != nil {
				t.Fatal(err)
			}
			if got := result.For(MarketScopeKR); got.Restored || got.Activation != nil || got.Reason != ResumeManifestMismatch {
				t.Fatalf("KR result=%+v", got)
			}
			if got := result.For(MarketScopeUS); !got.Restored || got.Activation == nil {
				t.Fatalf("US result=%+v", got)
			}
		})
	}
}

var pairedActivationNow = time.Date(2026, 8, 4, 4, 5, 6, 0, time.UTC)

func newProductionActivationFixture(t *testing.T) *productionActivationFixture {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &productionActivationFixture{dir: t.TempDir(), now: pairedActivationNow, public: public, private: private,
		configs: make(map[MarketScope]ProductionActivationConfig, 2), desired: make(map[MarketScope]DesiredState, 2), current: make(map[MarketScope]CurrentBinding, 2)}
	for index, market := range []MarketScope{MarketScopeKR, MarketScopeUS} {
		desired := activeDesired(market)
		current := currentBinding(desired)
		body := fixture.body(market, uint64(index+1), desired, current)
		data := signedActivationManifest(t, fixture.private, body)
		path := filepath.Join(fixture.dir, ProductionActivationFileName(market))
		if err := os.WriteFile(path, data, 0o400); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o400); err != nil {
			t.Fatal(err)
		}
		fixture.configs[market] = ProductionActivationConfig{ConfigDir: fixture.dir, Market: market,
			ManifestDigest: testActivationDigest(data), TrustedKeyID: "activation-key-1", TrustedKey: fixture.public}
		fixture.desired[market], fixture.current[market] = desired, current
	}
	return fixture
}

func (fixture *productionActivationFixture) requests(t *testing.T) [2]PairedRestoreRequest {
	t.Helper()
	out := [2]PairedRestoreRequest{}
	for index, market := range []MarketScope{MarketScopeKR, MarketScopeUS} {
		verifier, err := NewProductionActivationVerifier(fixture.configs[market])
		if err != nil {
			t.Fatalf("NewProductionActivationVerifier(%s): %v", market, err)
		}
		out[index] = PairedRestoreRequest{Market: market, Desired: fixture.desired[market], Current: fixture.current[market], Verifier: verifier}
	}
	return out
}

func (fixture *productionActivationFixture) rewrite(t *testing.T, market MarketScope, edit func(*testActivationManifestBody)) {
	t.Helper()
	desired, current := fixture.desired[market], fixture.current[market]
	body := fixture.body(market, map[MarketScope]uint64{MarketScopeKR: 1, MarketScopeUS: 2}[market], desired, current)
	edit(&body)
	data := signedActivationManifest(t, fixture.private, body)
	path := filepath.Join(fixture.dir, ProductionActivationFileName(market))
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	config := fixture.configs[market]
	config.ManifestDigest = testActivationDigest(data)
	fixture.configs[market] = config
}

func (fixture *productionActivationFixture) body(market MarketScope, generation uint64, desired DesiredState, current CurrentBinding) testActivationManifestBody {
	return testActivationManifestBody{SchemaVersion: testActivationSchema, Domain: testActivationDomain,
		SignatureAlgorithm: "Ed25519", KeyID: "activation-key-1", Generation: generation,
		SchedulerVersion: SchedulerVersion, DesiredRevision: desired.Revision, CalendarVersion: current.CalendarVersion,
		Market: market, Session: desired.Session, ConfigVersion: current.ConfigVersion, BuildDigest: current.BuildDigest,
		Actor: desired.Actor, ApprovedAt: desired.ApprovedAt.UTC().Format(time.RFC3339Nano),
		IssuedAt: fixture.now.Add(-time.Minute).Format(time.RFC3339Nano), ExpiresAt: fixture.now.Add(time.Hour).Format(time.RFC3339Nano)}
}

func signedActivationManifest(t *testing.T, private ed25519.PrivateKey, body testActivationManifestBody) []byte {
	t.Helper()
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := testActivationManifest{testActivationManifestBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func testActivationDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func activeDesired(market MarketScope) DesiredState {
	return DesiredState{Revision: 7, Version: SchedulerVersion, Enabled: true, AutoStart: true, Market: market,
		Session: SessionRegular, Actor: "human-approver", ApprovedAt: pairedActivationNow.Add(-2 * time.Minute),
		CalendarVersion: "calendar-v7-" + string(market), ConfigVersion: "strategy-config-v3"}
}

func currentBinding(desired DesiredState) CurrentBinding {
	return CurrentBinding{SchedulerVersion: SchedulerVersion, CalendarVersion: desired.CalendarVersion,
		Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion, BuildDigest: "sha256:" + strings.Repeat("b", 64)}
}

type blockingActivationVerifier struct {
	market  MarketScope
	started chan<- MarketScope
	release <-chan struct{}
}

func (verifier blockingActivationVerifier) verifyActivation(_ context.Context, binding ActivationBinding, now time.Time) (ActivationEvidence, error) {
	verifier.started <- verifier.market
	if verifier.release != nil {
		<-verifier.release
	}
	generation := binding.DesiredRevision
	if generation == 0 {
		generation = 1
	}
	return ActivationEvidence{Generation: generation, ExpiresAt: now.Add(time.Hour)}, nil
}

type countingActivationVerifier struct{ calls *atomic.Int32 }

func (verifier countingActivationVerifier) verifyActivation(_ context.Context, binding ActivationBinding, now time.Time) (ActivationEvidence, error) {
	verifier.calls.Add(1)
	generation := binding.DesiredRevision
	if generation == 0 {
		generation = 1
	}
	return ActivationEvidence{Generation: generation, ExpiresAt: now.Add(time.Hour)}, nil
}

func pairedActivationRequests(verifiers map[MarketScope]ActivationVerifier) [2]PairedRestoreRequest {
	return [2]PairedRestoreRequest{activationRequest(MarketScopeKR, verifiers[MarketScopeKR]), activationRequest(MarketScopeUS, verifiers[MarketScopeUS])}
}

func activationRequest(market MarketScope, verifier ActivationVerifier) PairedRestoreRequest {
	desired := activeDesired(market)
	return PairedRestoreRequest{Market: market, Desired: desired, Current: currentBinding(desired), Verifier: verifier}
}
