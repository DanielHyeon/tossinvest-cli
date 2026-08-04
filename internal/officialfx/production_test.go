package officialfx

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func TestProductionAuthorityCollectsPairedKRUS(t *testing.T) {
	fixture := newProductionAuthorityFixture(t, 1, "1.015")
	collection := fixture.service.Collect(context.Background())

	kr, err := collection.KR()
	if err != nil {
		t.Fatalf("KR: %v", err)
	}
	us, err := collection.US()
	if err != nil {
		t.Fatalf("US: %v", err)
	}
	if _, err := kr.EvidenceAt(fixture.now, "KRW", "KRW"); err != nil {
		t.Fatalf("KR evidence: %v", err)
	}
	reserve, err := us.EvidenceAt(fixture.now, "USD", "KRW")
	if err != nil {
		t.Fatalf("US evidence: %v", err)
	}
	if reserve.Haircut() != "1.015" {
		t.Fatalf("haircut = %q", reserve.Haircut())
	}
	if fixture.reader.accountCalls.Load() != 1 || fixture.reader.fxCalls.Load() != 1 {
		t.Fatalf("official calls: account=%d fx=%d", fixture.reader.accountCalls.Load(), fixture.reader.fxCalls.Load())
	}
}

func TestProductionAuthorityIsolatesKRFailure(t *testing.T) {
	fixture := newProductionAuthorityFixture(t, 1, "1.01")
	fixture.reader.accountErr = errors.New("account unavailable")
	collection := fixture.service.Collect(context.Background())
	if _, err := collection.KR(); !errors.Is(err, ErrProductionIdentityUnavailable) {
		t.Fatalf("KR error = %v", err)
	}
	if _, err := collection.US(); err != nil {
		t.Fatalf("US: %v", err)
	}
	if fixture.reader.accountCalls.Load() != 1 || fixture.reader.fxCalls.Load() != 1 {
		t.Fatalf("both paths not attempted: account=%d fx=%d", fixture.reader.accountCalls.Load(), fixture.reader.fxCalls.Load())
	}
}

func TestProductionAuthorityIsolatesUSFailure(t *testing.T) {
	fixture := newProductionAuthorityFixture(t, 1, "1.01")
	fixture.reader.fxErr = errors.New("fx unavailable")
	collection := fixture.service.Collect(context.Background())
	if _, err := collection.KR(); err != nil {
		t.Fatalf("KR: %v", err)
	}
	if _, err := collection.US(); !errors.Is(err, ErrProductionPolicyUnavailable) {
		t.Fatalf("US error = %v", err)
	}
	if fixture.reader.accountCalls.Load() != 1 || fixture.reader.fxCalls.Load() != 1 {
		t.Fatalf("both paths not attempted: account=%d fx=%d", fixture.reader.accountCalls.Load(), fixture.reader.fxCalls.Load())
	}
}

func TestProductionAuthorityMarketReadsNeverWaitForPeerPairedKRUS(t *testing.T) {
	t.Run("KR does not wait for blocked US", func(t *testing.T) {
		fixture := newProductionAuthorityFixture(t, 1, "1.01")
		started, release := make(chan struct{}, 1), make(chan struct{})
		fixture.reader.fxStarted, fixture.reader.fxRelease = started, release
		usDone := make(chan error, 1)
		go func() {
			_, err := fixture.service.CollectUS(context.Background())
			usDone <- err
		}()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("US read did not start")
		}
		krDone := make(chan error, 1)
		go func() {
			_, err := fixture.service.CollectKR(context.Background())
			krDone <- err
		}()
		select {
		case err := <-krDone:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("KR authority waited for blocked US")
		}
		close(release)
		if err := <-usDone; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("US does not wait for blocked KR", func(t *testing.T) {
		fixture := newProductionAuthorityFixture(t, 1, "1.01")
		started, release := make(chan struct{}, 1), make(chan struct{})
		fixture.reader.accountStarted, fixture.reader.accountRelease = started, release
		krDone := make(chan error, 1)
		go func() {
			_, err := fixture.service.CollectKR(context.Background())
			krDone <- err
		}()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("KR read did not start")
		}
		usDone := make(chan error, 1)
		go func() {
			_, err := fixture.service.CollectUS(context.Background())
			usDone <- err
		}()
		select {
		case err := <-usDone:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("US authority waited for blocked KR")
		}
		close(release)
		if err := <-krDone; err != nil {
			t.Fatal(err)
		}
	})
}

func TestProductionAuthorityUSManifestRefusalMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*riskPolicyManifest)
		remove bool
	}{
		{name: "absent", remove: true},
		{name: "unsigned", mutate: func(manifest *riskPolicyManifest) { manifest.Signature = "" }},
		{name: "stale", mutate: func(manifest *riskPolicyManifest) { manifest.FreshUntil = "2026-08-03T23:59:59Z" }},
		{name: "wrong account", mutate: func(manifest *riskPolicyManifest) { manifest.AccountID = "other" }},
		{name: "wrong base", mutate: func(manifest *riskPolicyManifest) { manifest.AccountCurrency = "USD" }},
		{name: "wrong quote", mutate: func(manifest *riskPolicyManifest) { manifest.QuoteCurrency = "EUR" }},
		{name: "wrong market", mutate: func(manifest *riskPolicyManifest) { manifest.Market = "KR" }},
		{name: "bad signature", mutate: func(manifest *riskPolicyManifest) {
			manifest.Signature = base64.StdEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newProductionAuthorityFixture(t, 1, "1.01")
			if test.remove {
				if err := os.Remove(filepath.Join(fixture.dir, RiskPolicyManifestFile)); err != nil {
					t.Fatal(err)
				}
			} else {
				manifest := fixture.manifest
				test.mutate(&manifest)
				fixture.writeManifest(t, manifest, false)
			}
			fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
			collection := fixture.service.Collect(context.Background())
			if _, err := collection.KR(); err != nil {
				t.Fatalf("KR: %v", err)
			}
			if _, err := collection.US(); err == nil {
				t.Fatal("US unexpectedly accepted invalid manifest")
			}
			if fixture.reader.fxCalls.Load() != 0 {
				t.Fatalf("FX read count = %d", fixture.reader.fxCalls.Load())
			}
		})
	}
}

func TestProductionAuthorityNeverDefaultsHaircut(t *testing.T) {
	for _, multiplier := range []string{"", "0", "0.99", "01.1", "1.0.1"} {
		t.Run(multiplier, func(t *testing.T) {
			fixture := newProductionAuthorityFixture(t, 1, "1.01")
			manifest := fixture.manifest
			manifest.Multiplier = multiplier
			fixture.writeManifest(t, manifest, true)
			fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
			collection := fixture.service.Collect(context.Background())
			if _, err := collection.KR(); err != nil {
				t.Fatalf("KR: %v", err)
			}
			if _, err := collection.US(); !errors.Is(err, ErrProductionPolicyInvalid) {
				t.Fatalf("US error = %v", err)
			}
			if fixture.reader.fxCalls.Load() != 0 {
				t.Fatalf("FX read count = %d", fixture.reader.fxCalls.Load())
			}
		})
	}
}

func TestProductionAuthorityRejectsNonCanonicalAndDuplicateManifest(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{name: "newline", mutate: func(data []byte) []byte { return append(data, '\n') }},
		{name: "unknown field", mutate: func(data []byte) []byte {
			return append(append([]byte(nil), data[:len(data)-1]...), []byte(`,"unexpected":true}`)...)
		}},
		{name: "duplicate field", mutate: func(data []byte) []byte {
			return []byte(`{"schema_version":"fx-risk-policy/v1",` + string(data[1:]))
		}},
		{name: "trailing value", mutate: func(data []byte) []byte { return append(append([]byte(nil), data...), []byte(`{}`)...) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newProductionAuthorityFixture(t, 1, "1.01")
			path := filepath.Join(fixture.dir, RiskPolicyManifestFile)
			canonical, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			data := test.mutate(canonical)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0o400); err != nil {
				t.Fatal(err)
			}
			fixture.digest = sha256Identity(string(data))
			fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
			collection := fixture.service.Collect(context.Background())
			if _, err := collection.KR(); err != nil {
				t.Fatalf("KR: %v", err)
			}
			if _, err := collection.US(); !errors.Is(err, ErrProductionPolicyInvalid) {
				t.Fatalf("US error = %v", err)
			}
			if fixture.reader.fxCalls.Load() != 0 {
				t.Fatalf("FX calls = %d", fixture.reader.fxCalls.Load())
			}
		})
	}
}

func TestProductionAuthorityRejectsTrustedTimeRollback(t *testing.T) {
	fixture := newProductionAuthorityFixture(t, 1, "1.01")
	if _, err := fixture.service.Collect(context.Background()).US(); err != nil {
		t.Fatal(err)
	}
	fixture.now = fixture.now.Add(-time.Second)
	fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
	collection := fixture.service.Collect(context.Background())
	if _, err := collection.KR(); err != nil {
		t.Fatalf("KR: %v", err)
	}
	if _, err := collection.US(); !errors.Is(err, ErrProductionAuthorityRollback) {
		t.Fatalf("US error = %v", err)
	}
}

func TestProductionAuthorityRejectsGenerationRollbackAndSubstitution(t *testing.T) {
	for _, test := range []struct {
		name       string
		generation uint64
		policyID   string
	}{
		{name: "generation rollback", generation: 1, policyID: "policy"},
		{name: "same generation substitution", generation: 2, policyID: "substituted"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newProductionAuthorityFixture(t, 2, "1.01")
			if _, err := fixture.service.Collect(context.Background()).US(); err != nil {
				t.Fatal(err)
			}
			manifest := fixture.manifest
			manifest.Generation = test.generation
			manifest.PolicyID = test.policyID
			fixture.writeManifest(t, manifest, true)
			fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
			collection := fixture.service.Collect(context.Background())
			if _, err := collection.KR(); err != nil {
				t.Fatalf("KR: %v", err)
			}
			if _, err := collection.US(); !errors.Is(err, ErrProductionAuthorityRollback) {
				t.Fatalf("US error = %v", err)
			}
		})
	}
}

func TestProductionAuthorityConcurrentExactReplay(t *testing.T) {
	fixture := newProductionAuthorityFixture(t, 1, "1.01")
	const collectors = 8
	var wait sync.WaitGroup
	errorsCh := make(chan error, collectors)
	for index := 0; index < collectors; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			service := newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
			_, err := service.Collect(context.Background()).US()
			errorsCh <- err
		}()
	}
	wait.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent replay: %v", err)
		}
	}
}

func TestProductionAuthorityCancelledContextMintsNothing(t *testing.T) {
	fixture := newProductionAuthorityFixture(t, 1, "1.01")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	collection := fixture.service.Collect(ctx)
	if _, err := collection.KR(); !errors.Is(err, ErrProductionAuthorityUnavailable) {
		t.Fatalf("KR error = %v", err)
	}
	if _, err := collection.US(); !errors.Is(err, ErrProductionAuthorityUnavailable) {
		t.Fatalf("US error = %v", err)
	}
	if fixture.reader.accountCalls.Load() != 0 || fixture.reader.fxCalls.Load() != 0 {
		t.Fatalf("cancelled collection performed reads: account=%d fx=%d", fixture.reader.accountCalls.Load(), fixture.reader.fxCalls.Load())
	}
}

func TestProductionAuthorityRejectsUnsafeManifestAndDeletedState(t *testing.T) {
	t.Run("manifest mode", func(t *testing.T) {
		fixture := newProductionAuthorityFixture(t, 1, "1.01")
		if err := os.Chmod(filepath.Join(fixture.dir, RiskPolicyManifestFile), 0o600); err != nil {
			t.Fatal(err)
		}
		collection := fixture.service.Collect(context.Background())
		if _, err := collection.KR(); err != nil {
			t.Fatalf("KR: %v", err)
		}
		if _, err := collection.US(); !errors.Is(err, ErrProductionPolicyUnavailable) {
			t.Fatalf("US error = %v", err)
		}
	})
	t.Run("state deletion after bootstrap", func(t *testing.T) {
		fixture := newProductionAuthorityFixture(t, 1, "1.01")
		if _, err := fixture.service.Collect(context.Background()).US(); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(fixture.dir, riskPolicyStateFile)); err != nil {
			t.Fatal(err)
		}
		fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
		collection := fixture.service.Collect(context.Background())
		if _, err := collection.KR(); err != nil {
			t.Fatalf("KR: %v", err)
		}
		if _, err := collection.US(); !errors.Is(err, ErrProductionAuthorityState) {
			t.Fatalf("US error = %v", err)
		}
	})
}

type fakeProductionOfficialReader struct {
	accountCalls   atomic.Int64
	fxCalls        atomic.Int64
	accountErr     error
	fxErr          error
	accountStarted chan<- struct{}
	accountRelease <-chan struct{}
	fxStarted      chan<- struct{}
	fxRelease      <-chan struct{}
}

func (reader *fakeProductionOfficialReader) reverifyAccount(ctx context.Context, _ string) error {
	reader.accountCalls.Add(1)
	if reader.accountStarted != nil {
		select {
		case reader.accountStarted <- struct{}{}:
		default:
		}
	}
	if reader.accountRelease != nil {
		select {
		case <-reader.accountRelease:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return reader.accountErr
}

func (reader *fakeProductionOfficialReader) readOfficial(ctx context.Context, quoteCurrency, accountCurrency string, policy HaircutPolicy) (Evidence, error) {
	reader.fxCalls.Add(1)
	if reader.fxStarted != nil {
		select {
		case reader.fxStarted <- struct{}{}:
		default:
		}
	}
	if reader.fxRelease != nil {
		select {
		case <-reader.fxRelease:
		case <-ctx.Done():
			return Evidence{}, ctx.Err()
		}
	}
	if reader.fxErr != nil {
		return Evidence{}, reader.fxErr
	}
	return sealOfficial(domain.ExchangeRate{Code: quoteCurrency + "/" + accountCurrency,
		BaseCurrency: quoteCurrency, QuoteCurrency: accountCurrency, RateRaw: "1380.5", MidRateRaw: "1375.25",
		ValidFromRaw: "2026-08-04T00:00:00Z", ValidUntilRaw: "2026-08-04T00:05:00Z"}, quoteCurrency, accountCurrency, policy)
}

type productionAuthorityFixture struct {
	dir      string
	now      time.Time
	public   ed25519.PublicKey
	private  ed25519.PrivateKey
	manifest riskPolicyManifest
	digest   string
	reader   *fakeProductionOfficialReader
	service  *ProductionAuthorityService
}

func newProductionAuthorityFixture(t *testing.T, generation uint64, multiplier string) *productionAuthorityFixture {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	seed := sha256.Sum256([]byte(t.Name()))
	private := ed25519.NewKeyFromSeed(seed[:])
	public := private.Public().(ed25519.PublicKey)
	fixture := &productionAuthorityFixture{dir: dir, now: time.Date(2026, 8, 4, 0, 1, 0, 0, time.UTC), public: public, private: private,
		reader: &fakeProductionOfficialReader{}}
	fixture.manifest = riskPolicyManifest{riskPolicyManifestBody: riskPolicyManifestBody{
		SchemaVersion: riskPolicySchema, AccountID: "7", AccountCurrency: "KRW", Market: "US", QuoteCurrency: "USD",
		Generation: generation, PolicyID: "policy", PolicyVersion: "v1", Multiplier: multiplier,
		ObservedAt: "2026-08-04T00:00:00Z", FreshUntil: "2026-08-04T00:05:00Z", Approver: "risk-committee",
		KeyID: "key-1", SignatureAlgorithm: riskPolicySignatureAlgorithm,
	}}
	fixture.writeManifest(t, fixture.manifest, true)
	fixture.service = newProductionAuthorityServiceWithReader(fixture.config(), fixture.reader)
	return fixture
}

func (fixture *productionAuthorityFixture) config() ProductionAuthorityConfig {
	return ProductionAuthorityConfig{ConfigDir: fixture.dir, AccountID: "7", AccountCurrency: "KRW",
		ManifestDigest: fixture.digest, TrustedKeyID: "key-1", TrustedKey: fixture.public,
		Now: func() time.Time { return fixture.now }}
}

func (fixture *productionAuthorityFixture) writeManifest(t *testing.T, manifest riskPolicyManifest, resign bool) {
	t.Helper()
	if resign {
		body, err := json.Marshal(manifest.riskPolicyManifestBody)
		if err != nil {
			t.Fatal(err)
		}
		manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(fixture.private, body))
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fixture.dir, RiskPolicyManifestFile)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	fixture.manifest = manifest
	fixture.digest = sha256Identity(string(data))
}
