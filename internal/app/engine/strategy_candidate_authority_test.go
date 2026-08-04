package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	candidatepkg "github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestStrategyCandidateAuthorityCollectsKRAndUSInOneWave(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	path := seedEmptyCandidateStore(t, dir, now)
	pins := map[StrategyMarket]string{}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		pins[market] = writeEngineThresholdFixture(t, dir, market, now)
	}
	loader := newStrategyCandidateAuthorityLoader(dir, path, func(key string) string {
		for market, pin := range pins {
			if key == strategyCandidateActivationDigestEnv(market) {
				return pin
			}
		}
		return ""
	})
	loader.openStore = func(ctx context.Context, opts candidatepkg.Options) (*candidatepkg.Store, error) {
		opts.FSProber = candidatepkg.FixedFSProber(candidatepkg.FSInfo{Name: "ext4"})
		return candidatepkg.OpenReadOnly(ctx, opts)
	}
	pair := loader.collect(context.Background(), readySchedulePair(now))
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		got := pair.Snapshot().For(market)
		if !got.Ready || got.Reason != StrategyCandidateReady || got.ThresholdVersion == "" ||
			got.ThresholdSetDigest == "" || got.EvidenceDigest == "" || got.ActivationDigest != pins[market] {
			t.Fatalf("candidate snapshot(%s) = %+v", market, got)
		}
	}
}

func TestStrategyCandidateAuthorityFailureIsMarketLocal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	path := seedEmptyCandidateStore(t, dir, now)
	pins := map[StrategyMarket]string{}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		pins[market] = writeEngineThresholdFixture(t, dir, market, now)
	}
	if err := os.Chmod(filepath.Join(dir, candidatepkg.ProductionThresholdActivationFileName(string(StrategyMarketKR))), 0o600); err != nil {
		t.Fatal(err)
	}
	loader := newStrategyCandidateAuthorityLoader(dir, path, func(key string) string {
		for market, pin := range pins {
			if key == strategyCandidateActivationDigestEnv(market) {
				return pin
			}
		}
		return ""
	})
	loader.openStore = func(ctx context.Context, opts candidatepkg.Options) (*candidatepkg.Store, error) {
		opts.FSProber = candidatepkg.FixedFSProber(candidatepkg.FSInfo{Name: "ext4"})
		return candidatepkg.OpenReadOnly(ctx, opts)
	}
	pair := loader.collect(context.Background(), readySchedulePair(now))
	if pair.Snapshot().KR.Ready || pair.Snapshot().KR.Reason != StrategyCandidateThresholdInvalid {
		t.Fatalf("KR snapshot = %+v", pair.Snapshot().KR)
	}
	if !pair.Snapshot().US.Ready || pair.Snapshot().US.Reason != StrategyCandidateReady {
		t.Fatalf("US snapshot = %+v", pair.Snapshot().US)
	}
}

func TestStrategyCandidateAuthorityScheduleOffPerformsNoAuthorityOrStoreRead(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	loader := newStrategyCandidateAuthorityLoader(t.TempDir(), filepath.Join(t.TempDir(), "missing.db"), func(string) string { return "" })
	var mu sync.Mutex
	loads, opens := 0, 0
	loader.loadAuthority = func(context.Context, candidatepkg.ProductionThresholdAuthorityConfig, time.Time, time.Duration) (candidatepkg.ProductionThresholdAuthority, error) {
		mu.Lock()
		loads++
		mu.Unlock()
		return candidatepkg.ProductionThresholdAuthority{}, fmt.Errorf("unexpected authority read")
	}
	loader.openStore = func(context.Context, candidatepkg.Options) (*candidatepkg.Store, error) {
		mu.Lock()
		opens++
		mu.Unlock()
		return nil, fmt.Errorf("unexpected store read")
	}
	pair := loader.collect(context.Background(), failedStrategySchedulePair(now))
	mu.Lock()
	defer mu.Unlock()
	if loads != 0 || opens != 0 {
		t.Fatalf("schedule OFF reads = authority:%d store:%d", loads, opens)
	}
	if pair.Snapshot().KR.Reason != StrategyCandidateScheduleNotReady || pair.Snapshot().US.Reason != StrategyCandidateScheduleNotReady {
		t.Fatalf("snapshots = KR:%+v US:%+v", pair.Snapshot().KR, pair.Snapshot().US)
	}
}

func readySchedulePair(now time.Time) strategyScheduleAuthorityPair {
	market := func(value StrategyMarket) strategyScheduleMarketAuthority {
		return strategyScheduleMarketAuthority{market: value, snapshot: StrategyScheduleMarketSnapshot{Market: value, Ready: true}}
	}
	return strategyScheduleAuthorityPair{observedAt: now, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}

func seedEmptyCandidateStore(t *testing.T, dir string, now time.Time) string {
	t.Helper()
	path := filepath.Join(dir, candidatepkg.DBFileName)
	store, err := candidatepkg.Open(context.Background(), candidatepkg.Options{Path: path, Clock: clock.NewFake(now),
		FSProber: candidatepkg.FixedFSProber(candidatepkg.FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeEngineThresholdFixture(t *testing.T, dir string, market StrategyMarket, approvedAt time.Time) string {
	t.Helper()
	marketName := string(market)
	evidence := []byte("engine synthetic threshold evidence " + marketName)
	evidenceDigest := candidatepkg.DigestEvidence(evidence)
	document := fmt.Sprintf(`{"version":"candidate-veto-engine-v1","market":%q,"session":"regular","metrics":[{"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},{"key":"extended","definition":"gain from stored first price","value":"50"},{"key":"near_high","definition":"distance below intraday high","value":"2.0"}],"sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-08-01T00:00:00Z"},"sample_count":100,"missing_rate":"0.1","evidence_digest":%q}`, marketName, evidenceDigest)
	setDigest, err := candidatepkg.DigestThresholdSetDocument(strings.NewReader(document), candidatepkg.ThresholdScope{Market: marketName, Session: candidatepkg.SessionRegular})
	if err != nil {
		t.Fatal(err)
	}
	activation := fmt.Sprintf(`{"version":"candidate-veto-engine-v1","market":%q,"session":"regular","set_digest":%q,"evidence_digest":%q,"approved_at":%q,"approved_by":"synthetic-human"}`,
		marketName, setDigest, evidenceDigest, approvedAt.Format(time.RFC3339))
	for path, data := range map[string][]byte{
		filepath.Join(dir, candidatepkg.ProductionThresholdSetFileName(marketName)):        []byte(document),
		filepath.Join(dir, candidatepkg.ProductionThresholdEvidenceFileName(marketName)):   evidence,
		filepath.Join(dir, candidatepkg.ProductionThresholdActivationFileName(marketName)): []byte(activation),
	} {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	sum := sha256.Sum256([]byte(activation))
	return "sha256:" + hex.EncodeToString(sum[:])
}
