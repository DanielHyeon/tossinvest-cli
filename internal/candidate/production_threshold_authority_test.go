package candidate

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
)

func TestProductionThresholdAuthorityLoadsExactKRAndUSFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	dir := t.TempDir()
	asOf := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	for _, market := range []string{MarketKR, MarketUS} {
		market := market
		t.Run(market, func(t *testing.T) {
			pin, setDigest, evidenceDigest := writeProductionThresholdFixture(t, dir, market, asOf)
			authority, err := LoadProductionThresholdAuthority(context.Background(), ProductionThresholdAuthorityConfig{
				ConfigDir: dir, Market: market, ActivationDigest: pin,
			}, asOf, 0)
			if err != nil {
				t.Fatalf("LoadProductionThresholdAuthority(%s): %v", market, err)
			}
			if !authority.Valid() || authority.Market() != market || authority.SetDigest() != setDigest ||
				authority.EvidenceDigest() != evidenceDigest || authority.ActivationDigest() != pin {
				t.Fatalf("authority(%s) = valid:%t market:%q set:%q evidence:%q activation:%q",
					market, authority.Valid(), authority.Market(), authority.SetDigest(), authority.EvidenceDigest(), authority.ActivationDigest())
			}
			if got := authority.ThresholdSet().VetoThresholds(); got.SeenLatePercentilePct != "80" || got.ExtendedGainPct != "50" || got.NearHighDistancePct != "2.0" {
				t.Fatalf("thresholds(%s) = %+v", market, got)
			}
		})
	}
}

func TestProductionThresholdAuthorityRejectsUnpinnedWrongModeAndCrossMarket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	asOf := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		mutate func(t *testing.T, dir, market, pin string) ProductionThresholdAuthorityConfig
	}{
		{"missing pin", func(t *testing.T, dir, market, pin string) ProductionThresholdAuthorityConfig {
			return ProductionThresholdAuthorityConfig{ConfigDir: dir, Market: market}
		}},
		{"wrong pin", func(t *testing.T, dir, market, pin string) ProductionThresholdAuthorityConfig {
			return ProductionThresholdAuthorityConfig{ConfigDir: dir, Market: market, ActivationDigest: "sha256:" + strings.Repeat("0", 64)}
		}},
		{"wrong mode", func(t *testing.T, dir, market, pin string) ProductionThresholdAuthorityConfig {
			if err := os.Chmod(filepath.Join(dir, ProductionThresholdActivationFileName(market)), 0o600); err != nil {
				t.Fatal(err)
			}
			return ProductionThresholdAuthorityConfig{ConfigDir: dir, Market: market, ActivationDigest: pin}
		}},
		{"cross market", func(t *testing.T, dir, market, pin string) ProductionThresholdAuthorityConfig {
			return ProductionThresholdAuthorityConfig{ConfigDir: dir, Market: MarketUS, ActivationDigest: pin}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			pin, _, _ := writeProductionThresholdFixture(t, dir, MarketKR, asOf)
			config := tc.mutate(t, dir, MarketKR, pin)
			got, err := LoadProductionThresholdAuthority(context.Background(), config, asOf, 0)
			if err == nil || got.Valid() {
				t.Fatalf("LoadProductionThresholdAuthority = (%+v, %v), want invalid refusal", got, err)
			}
		})
	}
}

func TestProductionThresholdAuthorityRejectsSymlinkAndTamperedEvidence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("production owner verification fails closed on windows")
	}
	asOf := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		mutate func(t *testing.T, dir string)
	}{
		{"activation symlink", func(t *testing.T, dir string) {
			path := filepath.Join(dir, ProductionThresholdActivationFileName(MarketKR))
			target := path + ".target"
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, data, 0o400); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
		}},
		{"tampered evidence", func(t *testing.T, dir string) {
			path := filepath.Join(dir, ProductionThresholdEvidenceFileName(MarketKR))
			if err := os.Chmod(path, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("tampered"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, 0o400); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			pin, _, _ := writeProductionThresholdFixture(t, dir, MarketKR, asOf)
			tc.mutate(t, dir)
			got, err := LoadProductionThresholdAuthority(context.Background(), ProductionThresholdAuthorityConfig{
				ConfigDir: dir, Market: MarketKR, ActivationDigest: pin,
			}, asOf, 0)
			if err == nil || got.Valid() {
				t.Fatalf("LoadProductionThresholdAuthority = (%+v, %v), want invalid refusal", got, err)
			}
		})
	}
}

func writeProductionThresholdFixture(t *testing.T, dir, market string, approvedAt time.Time) (string, string, string) {
	t.Helper()
	evidence := []byte("synthetic production threshold evidence for " + market)
	evidenceDigest := DigestEvidence(evidence)
	document := thresholdJSON(market, SessionRegular, evidenceDigest, "80")
	setDigest, err := DigestThresholdSetDocument(strings.NewReader(document), ThresholdScope{Market: market, Session: SessionRegular})
	if err != nil {
		t.Fatal(err)
	}
	activation := syntheticActivationJSON("candidate-veto-2026-07-31.1", market, SessionRegular, setDigest, evidenceDigest, approvedAt)
	for path, data := range map[string][]byte{
		filepath.Join(dir, ProductionThresholdSetFileName(market)):        []byte(document),
		filepath.Join(dir, ProductionThresholdEvidenceFileName(market)):   evidence,
		filepath.Join(dir, ProductionThresholdActivationFileName(market)): []byte(activation),
	} {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	sum := sha256.Sum256([]byte(activation))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:])), setDigest, evidenceDigest
}
