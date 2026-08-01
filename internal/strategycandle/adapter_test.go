package strategycandle_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycandle"
)

func officialPage(t *testing.T, adjusted bool) official.RawMinutePage {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/oauth2/token":
			_, _ = writer.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/candles":
			_, _ = writer.Write([]byte(`{"result":{"candles":[` +
				`{"timestamp":"2026-07-31T09:00:00+09:00","openPrice":"100.0100","highPrice":"101.100","lowPrice":"99.900","closePrice":"100.200","volume":"0.100","currency":"KRW"},` +
				`{"timestamp":"2026-07-31T09:01:00+09:00","openPrice":"100.2","highPrice":"101.1","lowPrice":"99.9","closePrice":"100.2","volume":"0.1","currency":"KRW"},` +
				`{"timestamp":"2026-07-31T09:02:00+09:00","openPrice":"100.2","highPrice":"101.1","lowPrice":"99.9","closePrice":"100.2","volume":"0.1","currency":"KRW"},` +
				`{"timestamp":"2026-07-31T09:03:00+09:00","openPrice":"100.2","highPrice":"101.1","lowPrice":"99.9","closePrice":"100.2","volume":"0.1","currency":"KRW"},` +
				`{"timestamp":"2026-07-31T09:04:00+09:00","openPrice":"100.2","highPrice":"101.1","lowPrice":"99.9","closePrice":"100.55","volume":"0.1","currency":"KRW"}` +
				`]}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)
	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(server.URL), official.WithHTTPClient(server.Client()))
	page, err := client.RawMinuteCandles(context.Background(), "KR", "005930", 5, "", adjusted)
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func TestOfficialDTOToStrategyMarketAdapterIsLosslessAndIdentityBound(t *testing.T) {
	now := time.Date(2026, 7, 31, 9, 5, 0, 0, time.FixedZone("KST", 9*3600))
	bar, err := strategycandle.AdaptAndSealClosedKRXFiveMinute("KR", "005930", officialPage(t, false), now)
	if err != nil {
		t.Fatal(err)
	}
	if !bar.Valid() || bar.Market() != "KR" || bar.Symbol() != "005930" ||
		bar.Source() != official.RawCandleSourceOfficialOpenAPI || bar.Adjusted() ||
		bar.Open() != "100.01" || bar.Close() != "100.55" || bar.Volume() != "0.5" {
		t.Fatalf("adapted bar = %+v", bar)
	}
	if _, err := strategycandle.AdaptAndSealClosedKRXFiveMinute("KR", "000660", officialPage(t, false), now); err == nil {
		t.Fatal("request/page identity mismatch accepted")
	}
}

func TestOfficialDTOAdapterRejectsZeroAndAdjustedPages(t *testing.T) {
	if _, err := strategycandle.AdaptOfficialMinutePage(official.RawMinutePage{}); err == nil {
		t.Fatal("zero/forged official page accepted")
	}
	if _, err := strategycandle.AdaptOfficialMinutePage(officialPage(t, true)); err == nil {
		t.Fatal("adjusted official page accepted")
	}
}
