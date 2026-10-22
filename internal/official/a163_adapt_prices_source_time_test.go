package official

// a163 tasks 2.5 — adaptPrices 는 브로커 `timestamp` 원문을 SourceTimestamp 에 그대로 보존함(design D2 · D3).
// 해독하지 않고 · 행을 버리지 않고 · 오류를 내지 않음. null · 키 없음은 디코드 규칙상 둘 다 "" 이고 우리 시계로 메우지 않음.
// 받은 시각(FetchedAt)은 그대로 채움.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestA163AdaptPricesKeepsTheBrokerTimeVerbatim(t *testing.T) {
	body := `[
		{"symbol":"005930","lastPrice":"72000","currency":"KRW","timestamp":"2026-08-27T19:59:59.000+09:00"},
		{"symbol":"AAPL","lastPrice":"185.70","currency":"USD","timestamp":"2026-08-28T05:28:30.000+09:00"},
		{"symbol":"000660","lastPrice":"120000","currency":"KRW","timestamp":null},
		{"symbol":"035420","lastPrice":"200000","currency":"KRW"},
		{"symbol":"051910","lastPrice":"9000","currency":"KRW","timestamp":"not-a-time"},
		{"symbol":"068270","lastPrice":"150000","currency":"KRW","timestamp":"2026-08-27T19:59:59.123+0900"}
	]`
	var raw []apiPrice
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	before := time.Now().UTC()
	got := adaptPrices(raw)
	after := time.Now().UTC()
	want := []string{
		"2026-08-27T19:59:59.000+09:00", "2026-08-28T05:28:30.000+09:00", "", "", "not-a-time", "2026-08-27T19:59:59.123+0900",
	}
	if len(got) != len(want) {
		t.Fatalf("adaptPrices kept %d rows, want %d — no row may be dropped for its timestamp", len(got), len(want))
	}
	for i, q := range got {
		if q.SourceTimestamp != want[i] {
			t.Errorf("%s SourceTimestamp = %q, want %q verbatim", q.Symbol, q.SourceTimestamp, want[i])
		}
		// 받은 시각은 그대로 받은 순간 — 브로커 시각으로 바꾸거나 유도하지 않음(spec official-read-compatibility MUST NOT).
		if q.FetchedAt.Before(before) || q.FetchedAt.After(after) {
			t.Errorf("%s FetchedAt = %s, want the receipt moment in [%s, %s]", q.Symbol, q.FetchedAt, before, after)
		}
	}
}

// 같은 보존을 HTTP 경로로 — Client.Prices 가 행 탈락 · 오류 없이 원문을 실어 나르고 /prices 호출은 한 번(rate budget 불변).
func TestA163PricesCarriesTheBrokerTimeInOneRequest(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/prices":
			calls++
			_, _ = w.Write([]byte(`{"result":[{"symbol":"005930","lastPrice":"72000","currency":"KRW","timestamp":"2026-08-27T19:59:59.000+09:00"},{"symbol":"000660","lastPrice":"1","currency":"KRW","timestamp":null}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	quotes, err := c.Prices(context.Background(), []string{"005930", "000660"})
	if err != nil {
		t.Fatalf("Prices: %v", err)
	}
	if calls != 1 || len(quotes) != 2 || quotes[0].SourceTimestamp != "2026-08-27T19:59:59.000+09:00" || quotes[1].SourceTimestamp != "" {
		t.Fatalf("calls=%d quotes=%+v, want one request and the broker time verbatim (null → empty)", calls, quotes)
	}
}
