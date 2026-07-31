package official

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestMarketCalendarReadsRequireExactGregorianDate(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"result":{}}`))
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	invalid := []string{
		"2026-7-03", "2026-07-3", "2026-02-29", "2026-02-30",
		" 2026-07-03", "2026-07-03 ", "2026-07-03T00:00:00Z",
	}
	for _, date := range invalid {
		t.Run(date, func(t *testing.T) {
			if _, err := c.TypedMarketCalendar(context.Background(), "KR", date); err == nil {
				t.Fatal("typed calendar accepted non-exact date")
			}
			if _, err := c.MarketCalendar(context.Background(), "KR", date); err == nil {
				t.Fatal("legacy calendar accepted non-exact date")
			}
		})
	}
	if requests != 0 {
		t.Fatalf("invalid dates reached official API %d time(s)", requests)
	}
}

func TestTypedMarketCalendarPreservesNullableSessions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/market-calendar/US":
			_, _ = w.Write([]byte(`{"result":{
				"previousBusinessDay":{"date":"2026-07-02"},
				"today":{"date":"2026-07-03","regularMarket":null},
				"nextBusinessDay":{"date":"2026-07-06","regularMarket":{
					"startTime":"2026-07-06T09:30:00-04:00","endTime":"2026-07-06T16:00:00-04:00"}}
			}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	got, err := c.TypedMarketCalendar(context.Background(), "us", "2026-07-03")
	if err != nil {
		t.Fatal(err)
	}
	if got.Today.Date != "2026-07-03" || got.Today.RegularMarket != nil {
		t.Fatalf("holiday = %+v", got.Today)
	}
	if got.NextBusinessDay.RegularMarket == nil || got.NextBusinessDay.RegularMarket.StartTime.IsZero() {
		t.Fatalf("next session = %+v", got.NextBusinessDay)
	}
}

func TestTypedMarketCalendarRejectsMalformedTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":{"today":{"date":"2026-03-25","regularMarket":{
			"startTime":"local-time-without-zone","endTime":"2026-03-25T16:00:00-04:00"}}}}`))
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	if _, err := c.TypedMarketCalendar(context.Background(), "US", "2026-03-25"); err == nil {
		t.Fatal("timezone-naive/malformed calendar time was accepted")
	}
}
