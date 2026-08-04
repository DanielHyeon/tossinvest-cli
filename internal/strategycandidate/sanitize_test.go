package strategycandidate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

func TestSealProducesOnlyAnImmutableKRAndUSStrategySnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	for _, market := range []string{candidate.MarketKR, candidate.MarketUS} {
		market := market
		t.Run(market, func(t *testing.T) {
			at := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
			authority := loadAuthority(t, market, at)
			first := at.Add(-time.Minute)
			last := at.Add(-30 * time.Second)
			verdict := candidate.Verdict{
				Summary: candidate.Summary{Candidate: candidate.Candidate{Key: candidate.Key{Market: market, Symbol: "SYNTHETIC"},
					FirstSeenAt: first, LastSeenAt: last, State: candidate.StateActive}},
				Sighting:  candidate.Sighting{Measured: true, Rank: 90, RankTotal: 100},
				Expansion: candidate.Expansion{Measured: true, FirstPrice: "100", LastPrice: "110", FirstAt: first, LastAt: last},
				Range:     candidate.RangePosition{Measured: true, High: "120", Price: "100", At: at},
			}
			approved, err := Seal(verdict, authority, at)
			if err != nil {
				t.Fatalf("Seal(%s): %v", market, err)
			}
			if !approved.Valid() || approved.Market() != market || approved.Symbol() != "SYNTHETIC" ||
				approved.SetDigest() != authority.SetDigest() || approved.EvidenceDigest() != authority.EvidenceDigest() {
				t.Fatalf("snapshot(%s) = valid:%t market:%q symbol:%q set:%q evidence:%q", market,
					approved.Valid(), approved.Market(), approved.Symbol(), approved.SetDigest(), approved.EvidenceDigest())
			}
		})
	}
}

func TestSealRefusesUnmeasuredKRAndUSWithoutSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	at := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	for _, market := range []string{candidate.MarketKR, candidate.MarketUS} {
		authority := loadAuthority(t, market, at)
		got, err := Seal(candidate.Verdict{Summary: candidate.Summary{Candidate: candidate.Candidate{
			Key: candidate.Key{Market: market, Symbol: "SYNTHETIC"}, FirstSeenAt: at.Add(-time.Minute),
			LastSeenAt: at.Add(-30 * time.Second), State: candidate.StateActive,
		}}}, authority, at)
		if err == nil || got.Valid() {
			t.Fatalf("Seal(%s) = (%+v, %v), want refusal", market, got, err)
		}
	}
}

func loadAuthority(t *testing.T, market string, at time.Time) candidate.ProductionThresholdAuthority {
	t.Helper()
	dir := t.TempDir()
	evidence := []byte("strategycandidate synthetic evidence " + market)
	evidenceDigest := candidate.DigestEvidence(evidence)
	document := fmt.Sprintf(`{"version":"candidate-veto-test-v1","market":%q,"session":"regular","metrics":[{"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},{"key":"extended","definition":"gain from stored first price","value":"50"},{"key":"near_high","definition":"distance below intraday high","value":"2.0"}],"sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-08-01T00:00:00Z"},"sample_count":100,"missing_rate":"0.1","evidence_digest":%q}`, market, evidenceDigest)
	setDigest, err := candidate.DigestThresholdSetDocument(strings.NewReader(document), candidate.ThresholdScope{Market: market, Session: candidate.SessionRegular})
	if err != nil {
		t.Fatal(err)
	}
	activation := fmt.Sprintf(`{"version":"candidate-veto-test-v1","market":%q,"session":"regular","set_digest":%q,"evidence_digest":%q,"approved_at":%q,"approved_by":"synthetic-human"}`,
		market, setDigest, evidenceDigest, at.Format(time.RFC3339))
	for path, data := range map[string][]byte{
		filepath.Join(dir, candidate.ProductionThresholdSetFileName(market)):        []byte(document),
		filepath.Join(dir, candidate.ProductionThresholdEvidenceFileName(market)):   evidence,
		filepath.Join(dir, candidate.ProductionThresholdActivationFileName(market)): []byte(activation),
	} {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	sum := sha256.Sum256([]byte(activation))
	authority, err := candidate.LoadProductionThresholdAuthority(context.Background(), candidate.ProductionThresholdAuthorityConfig{
		ConfigDir: dir, Market: market, ActivationDigest: "sha256:" + hex.EncodeToString(sum[:]),
	}, at, 0)
	if err != nil {
		t.Fatal(err)
	}
	return authority
}
