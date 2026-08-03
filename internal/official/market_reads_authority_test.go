package official

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestAdaptExchangeRatePreservesAuthorityFields(t *testing.T) {
	raw := apiExchangeRate{
		BaseCurrency: "USD", QuoteCurrency: "KRW", Rate: "1380.5000", MidRate: "1375.2500",
		ValidFrom: "2026-03-25T09:30:00+09:00", ValidUntil: "2026-03-25T09:31:00+09:00",
	}
	got := adaptExchangeRate(raw)
	if got.BaseCurrency != raw.BaseCurrency || got.QuoteCurrency != raw.QuoteCurrency ||
		got.RateRaw != raw.Rate || got.MidRateRaw != raw.MidRate ||
		got.ValidFromRaw != raw.ValidFrom || got.ValidUntilRaw != raw.ValidUntil {
		t.Fatalf("lossless exchange-rate fields: got %+v, raw %+v", got, raw)
	}
	if got.Base != 1375.25 || got.Close != 1380.5 {
		t.Fatalf("legacy display fields changed: base=%v close=%v", got.Base, got.Close)
	}
}

func TestExchangeRatePreservesAuthorityFieldsAcrossHTTPBoundary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/exchange-rate":
			_, _ = w.Write([]byte(`{"result":{"baseCurrency":"USD","quoteCurrency":"KRW","rate":"1380.5000","midRate":"1375.2500","validFrom":"2026-03-25T09:30:00+09:00","validUntil":"2026-03-25T09:31:00+09:00"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	client := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))

	got, err := client.ExchangeRate(context.Background(), "USD", "KRW")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseCurrency != "USD" || got.QuoteCurrency != "KRW" || got.RateRaw != "1380.5000" || got.MidRateRaw != "1375.2500" ||
		got.ValidFromRaw != "2026-03-25T09:30:00+09:00" || got.ValidUntilRaw != "2026-03-25T09:31:00+09:00" {
		t.Fatalf("lossless exchange-rate response: %+v", got)
	}
}
