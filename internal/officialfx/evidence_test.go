package officialfx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const validExchangeRateResult = `{"result":{"baseCurrency":"USD","quoteCurrency":"KRW","rate":"1380.5000","midRate":"1375.2500","basisPoint":"40","rateChangeType":"UP","validFrom":"2026-03-25T09:30:00+09:00","validUntil":"2026-03-25T09:31:00+09:00"}}`

func TestReadOfficialSealsLosslessEvidence(t *testing.T) {
	sealed, err := sealOfficial(validExchangeRate(), "USD", "KRW", mustHaircutPolicy(t, "1.0100"))
	if err != nil {
		t.Fatal(err)
	}
	if sealed.QuoteCurrency() != "USD" || sealed.AccountCurrency() != "KRW" || sealed.Digest() == "" {
		t.Fatalf("sealed identity: quote=%q account=%q digest=%q", sealed.QuoteCurrency(), sealed.AccountCurrency(), sealed.Digest())
	}

	at := time.Date(2026, 3, 25, 0, 30, 30, 0, time.UTC)
	fx, err := sealed.EvidenceAt(at, "USD", "KRW")
	if err != nil {
		t.Fatal(err)
	}
	if fx.RateQuoteToBase() != "1380.5000" || fx.Haircut() != "1.01" || fx.Source() != OfficialSource || fx.Version() != OfficialVersion ||
		fx.Digest() != sealed.Digest() {
		t.Fatalf("official FX evidence: %+v", fx)
	}
	if want := time.Date(2026, 3, 25, 0, 30, 0, 0, time.UTC); !fx.ObservedAt().Equal(want) {
		t.Fatalf("observed at: got %s want %s", fx.ObservedAt(), want)
	}
	if want := time.Date(2026, 3, 25, 0, 31, 0, 0, time.UTC); !fx.FreshUntil().Equal(want) {
		t.Fatalf("fresh until: got %s want %s", fx.FreshUntil(), want)
	}
	// The legacy float fields are intentionally not consulted. Authority is
	// minted from the exact raw decimal returned by the official endpoint.
}

func TestReadOfficialRefusesInvalidAuthority(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{name: "rate parse", payload: exchangeRatePayload("USD", "KRW", "1e3", "1375", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z")},
		{name: "mid rate parse", payload: exchangeRatePayload("USD", "KRW", "1380", "", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z")},
		{name: "valid from missing", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "", "2026-03-25T09:31:00Z")},
		{name: "valid until missing", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "2026-03-25T09:30:00Z", "")},
		{name: "validity parse", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "not-time", "2026-03-25T09:31:00Z")},
		{name: "validity reversed", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "2026-03-25T09:32:00Z", "2026-03-25T09:31:00Z")},
		{name: "pair mismatch", payload: exchangeRatePayload("USD", "EUR", "1380", "1375", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z")},
		{name: "noncanonical currency", payload: exchangeRatePayload("usd", "KRW", "1380", "1375", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rate := domain.ExchangeRate{
				Code: "USD/KRW", BaseCurrency: "USD", QuoteCurrency: "KRW", RateRaw: "1380",
				MidRateRaw: "1375", ValidFromRaw: "2026-03-25T09:30:00Z", ValidUntilRaw: "2026-03-25T09:31:00Z",
			}
			var envelope struct {
				Result struct {
					BaseCurrency, QuoteCurrency, Rate, MidRate, ValidFrom, ValidUntil string
				} `json:"result"`
			}
			if err := json.Unmarshal([]byte(test.payload), &envelope); err != nil {
				t.Fatal(err)
			}
			rate = domain.ExchangeRate{Code: envelope.Result.BaseCurrency + "/" + envelope.Result.QuoteCurrency,
				BaseCurrency: envelope.Result.BaseCurrency, QuoteCurrency: envelope.Result.QuoteCurrency,
				RateRaw: envelope.Result.Rate, MidRateRaw: envelope.Result.MidRate,
				ValidFromRaw: envelope.Result.ValidFrom, ValidUntilRaw: envelope.Result.ValidUntil}
			_, err := sealOfficial(rate, "USD", "KRW", mustHaircutPolicy(t, "1"))
			if !errors.Is(err, ErrInvalidEvidence) {
				t.Fatalf("error = %v, want ErrInvalidEvidence", err)
			}
		})
	}
	if _, err := ReadOfficial(context.Background(), nil, "USD", "KRW", HaircutPolicy{}); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("nil client error = %v", err)
	}
	if _, err := ReadOfficial(context.Background(), exchangeRateClient(t, validExchangeRateResult), "USD", "KRW", mustHaircutPolicy(t, "1")); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("configured origin error = %v", err)
	}
}

func TestEvidenceAtFailsClosedOutsideWindowAndAfterTamper(t *testing.T) {
	sealed, err := sealOfficial(validExchangeRate(), "USD", "KRW", mustHaircutPolicy(t, "1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []time.Time{
		{},
		time.Date(2026, 3, 25, 0, 29, 59, 0, time.UTC),
		time.Date(2026, 3, 25, 0, 31, 1, 0, time.UTC),
	} {
		if _, err := sealed.EvidenceAt(at, "USD", "KRW"); !errors.Is(err, ErrEvidenceNotCurrent) {
			t.Fatalf("at %s error = %v, want ErrEvidenceNotCurrent", at, err)
		}
	}
	tampered := sealed
	tampered.rate = "1"
	if _, err := tampered.EvidenceAt(time.Date(2026, 3, 25, 0, 30, 30, 0, time.UTC), "USD", "KRW"); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("tamper error = %v, want ErrInvalidEvidence", err)
	}
	if _, err := sealed.EvidenceAt(time.Date(2026, 3, 25, 0, 30, 30, 0, time.UTC), "EUR", "KRW"); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("pair substitution error = %v", err)
	}
}

func TestIdentitySealsSameCurrencyOnly(t *testing.T) {
	observed := time.Date(2026, 3, 25, 0, 30, 0, 0, time.UTC)
	fresh := observed.Add(time.Minute)
	snapshot, err := newIdentitySnapshot("KRW", "snapshot-kr-1", sha256Identity("snapshot-kr-1"), observed, fresh)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := Identity(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	fx, err := sealed.EvidenceAt(observed.Add(30*time.Second), "KRW", "KRW")
	if err != nil {
		t.Fatal(err)
	}
	if sealed.QuoteCurrency() != "KRW" || sealed.AccountCurrency() != "KRW" || fx.RateQuoteToBase() != "1" || fx.Haircut() != "1" ||
		fx.Source() != IdentitySource || fx.Version() != IdentityVersion {
		t.Fatalf("identity evidence: sealed=%+v fx=%+v", sealed, fx)
	}
	if _, err := Identity(IdentitySnapshot{}); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("zero snapshot error = %v", err)
	}
	if _, err := newIdentitySnapshot("KRW", strings.Repeat("x", 257), sha256Identity("x"), observed, fresh); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("oversized identity error = %v", err)
	}
	if _, err := newIdentitySnapshot("KRW", "wide", sha256Identity("wide"), observed, observed.Add(maxIdentityWindow+time.Nanosecond)); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("unbounded snapshot error = %v", err)
	}
	tampered := snapshot
	tampered.freshUntil = fresh.Add(time.Minute)
	if _, err := Identity(tampered); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("tampered snapshot error = %v", err)
	}
}

func TestHaircutPolicyIsSealedCanonicalAndNotCallerMinted(t *testing.T) {
	observed := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	policy, err := newHaircutPolicy("fx-haircut-1", "policy-v1", "1.0100", observed, observed.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if policy.multiplier != "1.01" || policy.digest == "" || !policy.validAt(observed.Add(time.Hour)) {
		t.Fatalf("canonical policy=%+v", policy)
	}
	for _, raw := range []string{"0.99", "1e0", "01.01", "1.", strings.Repeat("9", 129)} {
		if _, err := newHaircutPolicy("fx-haircut-1", "policy-v1", raw, observed, observed.Add(time.Hour)); !errors.Is(err, ErrInvalidEvidence) {
			t.Fatalf("haircut %q error=%v", raw, err)
		}
	}
}

func validExchangeRate() domain.ExchangeRate {
	return domain.ExchangeRate{Code: "USD/KRW", BaseCurrency: "USD", QuoteCurrency: "KRW",
		RateRaw: "1380.5000", MidRateRaw: "1375.2500",
		ValidFromRaw: "2026-03-25T09:30:00+09:00", ValidUntilRaw: "2026-03-25T09:31:00+09:00"}
}

func mustHaircutPolicy(t *testing.T, multiplier string) HaircutPolicy {
	t.Helper()
	observed := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	policy, err := newHaircutPolicy("fx-haircut-1", "policy-v1", multiplier, observed, observed.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func exchangeRateClient(t *testing.T, payload string) *official.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/exchange-rate":
			_, _ = w.Write([]byte(payload))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()))
}

func exchangeRatePayload(base, quote, rate, mid, validFrom, validUntil string) string {
	return `{"result":{"baseCurrency":"` + base + `","quoteCurrency":"` + quote + `","rate":"` + rate + `","midRate":"` + mid +
		`","validFrom":"` + validFrom + `","validUntil":"` + validUntil + `"}}`
}
