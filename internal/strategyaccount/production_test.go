//go:build tossos_testseams

package strategyaccount

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

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

func TestProductionAccountAuthorityPairedKRUSSameWave(t *testing.T) {
	now := time.Date(2026, 8, 4, 3, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		market := market
		t.Run(string(market), func(t *testing.T) {
			symbol, quote, cash := "005930", "KRW", "5000000"
			if market == MarketUS {
				symbol, quote, cash = "AAPL", "USD", "1000.25"
			}
			body := accountBodyFixture(now, market, symbol, quote, cash)
			manifestDigest := writeAccountManifest(t, dir, body, private)
			authority, err := LoadProductionAuthority(context.Background(), ProductionConfig{ConfigDir: dir, AccountRef: "acct-7",
				AccountCurrency: "KRW", Symbol: symbol, Market: market, ManifestDigest: manifestDigest, TrustedKeyID: "account-key-1",
				TrustedKey: public, ObservedAt: now})
			if err != nil {
				t.Fatal(err)
			}
			account := authority.AccountState()
			if authority.Market() != market || authority.QuoteCurrency() != quote || authority.ManifestDigest() != manifestDigest ||
				authority.Identity() == "" || account.Mode != risk.ModeNormal || account.CashAvailable.Amount != cash ||
				account.OpenExposure.Amount != "100000" || len(account.AllowedSymbols) != 1 || account.AllowedSymbols[0] != symbol {
				t.Fatalf("authority=%+v account=%+v", authority, account)
			}
			account.AllowedSymbols[0] = "MUTATED"
			if authority.AccountState().AllowedSymbols[0] != symbol {
				t.Fatal("returned account slice mutated sealed authority")
			}
		})
	}
}

func TestProductionAccountAuthorityRefusesMarketLocalTamper(t *testing.T) {
	now := time.Date(2026, 8, 4, 3, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	body := accountBodyFixture(now, MarketKR, "005930", "KRW", "5000000")
	digest := writeAccountManifest(t, dir, body, private)
	if err := os.Chmod(filepath.Join(dir, FileName(MarketKR)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProductionAuthority(context.Background(), ProductionConfig{ConfigDir: dir, AccountRef: "acct-7", AccountCurrency: "KRW",
		Symbol: "005930", Market: MarketKR, ManifestDigest: digest, TrustedKeyID: "account-key-1", TrustedKey: public, ObservedAt: now}); err == nil {
		t.Fatal("owner-only mode tamper accepted")
	}
}

func accountBodyFixture(now time.Time, market Market, symbol, quote, cash string) productionBody {
	return productionBody{SchemaVersion: productionSchema, Domain: productionDomain, SignatureAlgorithm: productionAlgorithm,
		KeyID: "account-key-1", Generation: 7, Market: market, AccountRef: "acct-7", AccountCurrency: "KRW", QuoteCurrency: quote,
		Source: "toss-open-api", SourceDigest: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", Official: true,
		ObservedAt: now.Add(-time.Second).Format(time.RFC3339Nano), FreshUntil: now.Add(time.Second).Format(time.RFC3339Nano),
		OperatingMode: string(risk.ModeNormal), AllowedSymbols: []string{symbol}, HeldQuantity: "0", CashAvailable: cash,
		OpenExposureBase: "100000", DailyRealizedLossBase: "0", AccountEquityBase: "10000000"}
}

func writeAccountManifest(t *testing.T, dir string, body productionBody, private ed25519.PrivateKey) string {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionManifest{productionBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, FileName(body.Market))
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	return digest(data)
}
