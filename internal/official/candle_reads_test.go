package official

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// TestAdaptCandlesUnit verifies the pure adapter for Candles.
func TestAdaptCandlesUnit(t *testing.T) {
	raw := apiCandlePage{
		Candles: []apiCandle{
			{
				Timestamp:  "2026-03-25T09:00:00+09:00",
				OpenPrice:  "71600",
				HighPrice:  "72300",
				LowPrice:   "71500",
				ClosePrice: "72000",
				Volume:     "3521000",
				Currency:   "KRW",
			},
		},
		NextBefore: "2026-03-25T08:59:59+09:00",
	}

	got := adaptCandles("005930", "1d", raw)

	if got.Symbol != "005930" {
		t.Fatalf("Symbol: want 005930, got %q", got.Symbol)
	}
	if got.Interval != "1d" {
		t.Fatalf("Interval: want 1d, got %q", got.Interval)
	}
	if len(got.Candles) != 1 {
		t.Fatalf("len(Candles): want 1, got %d", len(got.Candles))
	}
	c := got.Candles[0]
	if c.Open != 71600 {
		t.Fatalf("Open: want 71600, got %v", c.Open)
	}
	if c.High != 72300 {
		t.Fatalf("High: want 72300, got %v", c.High)
	}
	if c.Low != 71500 {
		t.Fatalf("Low: want 71500, got %v", c.Low)
	}
	if c.Close != 72000 {
		t.Fatalf("Close: want 72000, got %v", c.Close)
	}
	if c.Volume != 3521000 {
		t.Fatalf("Volume: want 3521000, got %v", c.Volume)
	}
	// Timestamp must be parsed to a non-zero time.
	if c.Time.IsZero() {
		t.Fatal("Time: expected non-zero")
	}
	// The parsed time should be 2026-03-25 00:00:00 UTC (09:00 KST = UTC+9).
	want := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !c.Time.UTC().Equal(want) {
		t.Fatalf("Time UTC: want %v, got %v", want, c.Time.UTC())
	}
	// Fields not available from /candles endpoint.
	if got.ProductCode != "" {
		t.Fatalf("ProductCode: expected empty, got %q", got.ProductCode)
	}
	if got.Name != "" {
		t.Fatalf("Name: expected empty, got %q", got.Name)
	}
	if got.Base != 0 {
		t.Fatalf("Base: expected 0, got %v", got.Base)
	}
}

// TestAdaptCandlesEmpty verifies empty candle slice handling.
func TestAdaptCandlesEmpty(t *testing.T) {
	got := adaptCandles("AAPL", "1m", apiCandlePage{})
	if got.Symbol != "AAPL" {
		t.Fatalf("Symbol: want AAPL, got %q", got.Symbol)
	}
	if len(got.Candles) != 0 {
		t.Fatalf("Candles: want empty, got %d", len(got.Candles))
	}
}

func TestAdaptCandlesOrdersNewestFirstWireChronologically(t *testing.T) {
	raw := apiCandlePage{Candles: []apiCandle{
		{Timestamp: "2026-03-25T09:03:00+09:00", OpenPrice: "301", HighPrice: "302", LowPrice: "300", ClosePrice: "301.5", Volume: "3000"},
		{Timestamp: "2026-03-25T09:02:00+09:00", OpenPrice: "201", HighPrice: "202", LowPrice: "200", ClosePrice: "201.5", Volume: "2000"},
		{Timestamp: "2026-03-25T09:01:00+09:00", OpenPrice: "101", HighPrice: "102", LowPrice: "100", ClosePrice: "101.5", Volume: "1000"},
	}}
	got := adaptCandles("005930", "1m", raw)
	parse := func(value string) time.Time {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	want := []domain.Candle{
		{Time: parse("2026-03-25T09:01:00+09:00"), Open: 101, High: 102, Low: 100, Close: 101.5, Volume: 1000},
		{Time: parse("2026-03-25T09:02:00+09:00"), Open: 201, High: 202, Low: 200, Close: 201.5, Volume: 2000},
		{Time: parse("2026-03-25T09:03:00+09:00"), Open: 301, High: 302, Low: 300, Close: 301.5, Volume: 3000},
	}
	if !reflect.DeepEqual(got.Candles, want) {
		t.Fatalf("candles = %+v, want %+v", got.Candles, want)
	}
}

// TestAdaptCandlesBadTimestamp verifies that an unparseable timestamp yields
// a zero time.Time rather than a hard error.
func TestAdaptCandlesBadTimestamp(t *testing.T) {
	raw := apiCandlePage{
		Candles: []apiCandle{
			{Timestamp: "not-a-date", ClosePrice: "100"},
		},
	}
	got := adaptCandles("X", "1d", raw)
	if !got.Candles[0].Time.IsZero() {
		t.Fatalf("Time: expected zero for bad timestamp, got %v", got.Candles[0].Time)
	}
}

// TestCandlesIntegration tests Candles() against an httptest server.
func TestCandlesIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/candles":
			q := r.URL.Query()
			if q.Get("symbol") != "005930" {
				t.Errorf("symbol: want 005930, got %q", q.Get("symbol"))
			}
			if q.Get("interval") != "1d" {
				t.Errorf("interval: want 1d, got %q", q.Get("interval"))
			}
			_, _ = w.Write([]byte(`{"result":{"candles":[{"timestamp":"2026-03-25T09:00:00+09:00","openPrice":"71600","highPrice":"72300","lowPrice":"71500","closePrice":"72000","volume":"3521000","currency":"KRW"}],"nextBefore":"2026-03-25T08:59:59+09:00"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(
		Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
	)

	got, err := c.Candles(context.Background(), "005930", "1d", 0, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Symbol != "005930" {
		t.Fatalf("Symbol: want 005930, got %q", got.Symbol)
	}
	if len(got.Candles) != 1 {
		t.Fatalf("Candles: want 1, got %d", len(got.Candles))
	}
	if got.Candles[0].Close != 72000 {
		t.Fatalf("Close: want 72000, got %v", got.Candles[0].Close)
	}
}
