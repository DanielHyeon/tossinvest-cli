package officialfx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const validExchangeRateResult = `{"result":{"baseCurrency":"USD","quoteCurrency":"KRW","rate":"1380.5000","midRate":"1375.2500","basisPoint":"40","rateChangeType":"UP","validFrom":"2026-03-25T09:30:00+09:00","validUntil":"2026-03-25T09:31:00+09:00"}}`

func TestReadOfficialSealsLosslessEvidence(t *testing.T) {
	client := exchangeRateClient(t, validExchangeRateResult)
	sealed, err := ReadOfficial(context.Background(), client, "USD", "KRW", "1.0100")
	if err != nil {
		t.Fatal(err)
	}
	if sealed.QuoteCurrency() != "USD" || sealed.AccountCurrency() != "KRW" || sealed.Digest() == "" {
		t.Fatalf("sealed identity: quote=%q account=%q digest=%q", sealed.QuoteCurrency(), sealed.AccountCurrency(), sealed.Digest())
	}

	at := time.Date(2026, 3, 25, 0, 30, 30, 0, time.UTC)
	fx, err := sealed.EvidenceAt(at)
	if err != nil {
		t.Fatal(err)
	}
	if fx.RateQuoteToBase != "1380.5000" || fx.Haircut != "1.0100" || fx.Source != OfficialSource || fx.Version != OfficialVersion ||
		fx.Digest != sealed.Digest() || !fx.Official || !fx.Frozen {
		t.Fatalf("official FX evidence: %+v", fx)
	}
	if want := time.Date(2026, 3, 25, 0, 30, 0, 0, time.UTC); !fx.ObservedAt.Equal(want) {
		t.Fatalf("observed at: got %s want %s", fx.ObservedAt, want)
	}
	if want := time.Date(2026, 3, 25, 0, 31, 0, 0, time.UTC); !fx.FreshUntil.Equal(want) {
		t.Fatalf("fresh until: got %s want %s", fx.FreshUntil, want)
	}
	// The legacy float fields are intentionally not consulted. Authority is
	// minted from the exact raw decimal returned by the official endpoint.
}

func TestReadOfficialRefusesInvalidAuthority(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		haircut string
	}{
		{name: "rate parse", payload: exchangeRatePayload("USD", "KRW", "1e3", "1375", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "mid rate parse", payload: exchangeRatePayload("USD", "KRW", "1380", "", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "valid from missing", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "valid until missing", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "2026-03-25T09:30:00Z", ""), haircut: "1"},
		{name: "validity parse", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "not-time", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "validity reversed", payload: exchangeRatePayload("USD", "KRW", "1380", "1375", "2026-03-25T09:32:00Z", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "pair mismatch", payload: exchangeRatePayload("USD", "EUR", "1380", "1375", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "noncanonical currency", payload: exchangeRatePayload("usd", "KRW", "1380", "1375", "2026-03-25T09:30:00Z", "2026-03-25T09:31:00Z"), haircut: "1"},
		{name: "haircut below one", payload: validExchangeRateResult, haircut: "0.99"},
		{name: "haircut parse", payload: validExchangeRateResult, haircut: "1e0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ReadOfficial(context.Background(), exchangeRateClient(t, test.payload), "USD", "KRW", test.haircut)
			if !errors.Is(err, ErrInvalidEvidence) {
				t.Fatalf("error = %v, want ErrInvalidEvidence", err)
			}
		})
	}
	if _, err := ReadOfficial(context.Background(), nil, "USD", "KRW", "1"); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("nil client error = %v", err)
	}
}

func TestEvidenceAtFailsClosedOutsideWindowAndAfterTamper(t *testing.T) {
	sealed, err := ReadOfficial(context.Background(), exchangeRateClient(t, validExchangeRateResult), "USD", "KRW", "1")
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []time.Time{
		{},
		time.Date(2026, 3, 25, 0, 29, 59, 0, time.UTC),
		time.Date(2026, 3, 25, 0, 31, 1, 0, time.UTC),
	} {
		if _, err := sealed.EvidenceAt(at); !errors.Is(err, ErrEvidenceNotCurrent) {
			t.Fatalf("at %s error = %v, want ErrEvidenceNotCurrent", at, err)
		}
	}
	tampered := sealed
	tampered.rate = "1"
	if _, err := tampered.EvidenceAt(time.Date(2026, 3, 25, 0, 30, 30, 0, time.UTC)); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("tamper error = %v, want ErrInvalidEvidence", err)
	}
}

func TestIdentitySealsSameCurrencyOnly(t *testing.T) {
	observed := time.Date(2026, 3, 25, 0, 30, 0, 0, time.UTC)
	fresh := observed.Add(time.Minute)
	sealed, err := Identity("KRW", observed, fresh)
	if err != nil {
		t.Fatal(err)
	}
	fx, err := sealed.EvidenceAt(observed.Add(30 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if sealed.QuoteCurrency() != "KRW" || sealed.AccountCurrency() != "KRW" || fx.RateQuoteToBase != "1" || fx.Haircut != "1" ||
		fx.Source != IdentitySource || fx.Version != IdentityVersion || !fx.Official || !fx.Frozen {
		t.Fatalf("identity evidence: sealed=%+v fx=%+v", sealed, fx)
	}
	for _, test := range []struct {
		currency string
		from     time.Time
		until    time.Time
	}{
		{currency: "", from: observed, until: fresh},
		{currency: "krw", from: observed, until: fresh},
		{currency: "KRW", from: time.Time{}, until: fresh},
		{currency: "KRW", from: observed, until: time.Time{}},
		{currency: "KRW", from: fresh, until: observed},
	} {
		if _, err := Identity(test.currency, test.from, test.until); !errors.Is(err, ErrInvalidEvidence) {
			t.Fatalf("Identity(%q,%s,%s) error = %v", test.currency, test.from, test.until, err)
		}
	}
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
