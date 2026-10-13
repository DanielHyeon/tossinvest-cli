package official

// a130_intake_candle_test.go 는 a130 의 공식 1분봉 하나 읽기를 httptest 로 잼 — 실 endpoint 없음.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func a130CandleServer(t *testing.T, status int, body string) (*Client, *[]string) {
	t.Helper()
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/candles":
			queries = append(queries, r.URL.RawQuery)
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client())), &queries
}

func a130Bar(timestamp, closePrice, nextBefore string) string {
	return `{"result":{"candles":[{"timestamp":"` + timestamp + `","openPrice":"18540","highPrice":"18600","lowPrice":"18420",` +
		`"closePrice":"` + closePrice + `","volume":"275728","currency":"KRW"}],"nextBefore":"` + nextBefore + `"}}`
}

// TestA130TheMinuteCandleIsReadUnadjustedAndOnlyForItsOwnMinute: 요청은 1m · count 1 · before = 봉 이름(서울 오프셋) · adjusted=false 이고,
// 돌려받은 봉의 이름이 그 분일 때만 종가(소수 문자열 그대로)와 그 이름 — 더 이른 봉(그 분에 체결이 없을 때 API 가 주는 것)은 ok=false 이고
// 그 이른 이름을 돌려줌.
func TestA130TheMinuteCandleIsReadUnadjustedAndOnlyForItsOwnMinute(t *testing.T) {
	label := time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)
	client, queries := a130CandleServer(t, http.StatusOK, a130Bar("2026-10-07T09:19:00+09:00", "18430", "2026-10-07T09:18:00+09:00"))
	got, bar, ok, err := client.MinuteCandleClosingAt(context.Background(), "kr", "010170", label)
	if err != nil || !ok || got != "18430" || !bar.Equal(label) {
		t.Fatalf("got %q bar %v ok=%v err=%v, want 18430 at the requested minute", got, bar, ok, err)
	}
	want := "adjusted=false&before=2026-10-07T09%3A19%3A00%2B09%3A00&count=1&interval=1m&symbol=010170"
	if len(*queries) != 1 || (*queries)[0] != want {
		t.Errorf("query = %v, want %s", *queries, want)
	}

	earlier, _ := a130CandleServer(t, http.StatusOK, a130Bar("2026-10-07T09:18:00+09:00", "18540", "2026-10-07T09:17:00+09:00"))
	got, bar, ok, err = earlier.MinuteCandleClosingAt(context.Background(), "kr", "010170", label)
	if err != nil || ok || got != "" || !bar.Equal(label.Add(-time.Minute)) {
		t.Errorf("an earlier bar: got %q bar %v ok=%v err=%v, want no price and the earlier name", got, bar, ok, err)
	}
	empty, _ := a130CandleServer(t, http.StatusOK, `{"result":{"candles":[],"nextBefore":null}}`)
	if _, _, ok, err := empty.MinuteCandleClosingAt(context.Background(), "kr", "010170", label); err != nil || ok {
		t.Errorf("no bar: ok=%v err=%v, want no price", ok, err)
	}
}

// TestA130ARateLimitedCandleReadFailsAfterOneRequest: 429 는 오류이고 재시도하지 않음(도구는 쓰기 전에 멈춤 — 공유 쿼터를 더 쓰지 않음).
func TestA130ARateLimitedCandleReadFailsAfterOneRequest(t *testing.T) {
	client, queries := a130CandleServer(t, http.StatusTooManyRequests, `{"error":{"code":"RATE_LIMITED"}}`)
	if _, _, ok, err := client.MinuteCandleClosingAt(context.Background(), "kr", "010170",
		time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)); err == nil || ok {
		t.Fatalf("ok=%v err=%v, want an error", ok, err)
	}
	if len(*queries) != 1 {
		t.Errorf("requests = %d, want exactly one (no retry)", len(*queries))
	}
}
