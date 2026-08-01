package official

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestRawMinuteCandlesPreservesOfficialDecimalAndTimestampStrings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/candles":
			if r.URL.Query().Get("interval") != "1m" || r.URL.Query().Get("symbol") != "005930" {
				t.Errorf("query=%s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"result":{"candles":[{"timestamp":"2026-07-31T09:00:00+09:00","openPrice":"100.0100","highPrice":"101.100","lowPrice":"99.900","closePrice":"100.200","volume":"0.100","currency":"KRW"}],"nextBefore":"cursor"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	client := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"), WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	got, err := client.RawMinuteCandles(context.Background(), "KR", "005930", 5, "", false)
	if err != nil {
		t.Fatal(err)
	}
	candles := got.Candles()
	if !got.Valid() || got.Market() != "KR" || got.Symbol() != "005930" || got.Interval() != RawCandleIntervalOneMinute ||
		got.Adjusted() || got.Source() != RawCandleSourceOfficialOpenAPI || len(candles) != 1 ||
		candles[0].Open != "100.0100" || candles[0].Volume != "0.100" ||
		candles[0].Timestamp != "2026-07-31T09:00:00+09:00" || got.NextBefore() != "cursor" {
		t.Fatalf("raw=%+v", got)
	}
	candles[0].Open = "forged"
	if got.Candles()[0].Open != "100.0100" {
		t.Fatal("caller mutated opaque official page through candle slice")
	}
	if (RawMinutePage{}).Valid() {
		t.Fatal("zero official page accepted")
	}
}

func TestRawMinuteCandlesRejectsUnsupportedMarketAndAdjustedStrategyReadRemainsExplicit(t *testing.T) {
	client := &Client{}
	if _, err := client.RawMinuteCandles(context.Background(), "US", "AAPL", 5, "", false); err == nil {
		t.Fatal("non-KRX raw strategy market accepted")
	}
}
