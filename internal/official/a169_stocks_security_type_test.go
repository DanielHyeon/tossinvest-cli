package official

// a169 tasks 2.1 — 공식 /stocks 의 `securityType` 을 tossctl 출력 `security_type` 으로 드러냄(사용자 결정 2026-10-23 · Q11).
// 예정 소비자(a169 종목군 유형 갈래 · a128 D2 유형 영수증 — 2026-10-23 기준 읽는 코드 0)는 tossctl 출력 JSON 의 `security_type`
// 또는 `Client.Stocks` 가 돌려준 Quote 의 Go 필드를 읽음. 단언은 바깥 계약인 출력 JSON(`tossctl quote --json` 과 같은 writer)으로 함.
// 값이 없는 행(키 누락 · 빈 문자열)은 키 자체가 없어야 함 — omitempty 로 편집 전 바이트를 지킴(안전 3 · a163 golden 이 같은 것을 고정).

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
)

// a169StocksETFQuoteGolden 은 securityType 이 실린 /stocks 행의 출력 바이트임 — 키 위치(맨 끝)까지 고정.
const a169StocksETFQuoteGolden = "{\n  \"symbol\": \"069500\",\n  \"name\": \"KODEX 200\",\n  \"market_code\": \"KOSPI\",\n  \"currency\": \"KRW\",\n  \"status\": \"ACTIVE\",\n  \"fetched_at\": \"2026-10-23T00:00:00Z\",\n  \"security_type\": \"ETF\"\n}\n"

// 브로커 응답 그대로의 httptest — ETF · ETN · STOCK 은 값 원문, 마지막 행은 securityType 키가 아예 없음.
func TestA169StocksCarriesTheBrokerSecurityTypeVerbatim(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/stocks":
			_, _ = w.Write([]byte(`{"result":[` +
				`{"symbol":"069500","name":"KODEX 200","market":"KOSPI","securityType":"ETF","status":"ACTIVE","currency":"KRW"},` +
				`{"symbol":"520001","name":"ETN 예시","market":"KR_ETC","securityType":"ETN","status":"ACTIVE","currency":"KRW"},` +
				`{"symbol":"005930","name":"삼성전자","market":"KOSPI","securityType":"STOCK","status":"ACTIVE","currency":"KRW"},` +
				`{"symbol":"000000","name":"유형 없음","market":"KOSPI","status":"ACTIVE","currency":"KRW"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	quotes, err := c.Stocks(context.Background(), []string{"069500", "520001", "005930", "000000"})
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		symbol  string
		present bool
		value   string
	}{{"069500", true, "ETF"}, {"520001", true, "ETN"}, {"005930", true, "STOCK"}, {"000000", false, ""}}
	if len(quotes) != len(want) {
		t.Fatalf("Stocks = %d quotes, want %d", len(quotes), len(want))
	}
	for i, w := range want {
		var buf bytes.Buffer
		if err := output.WriteQuote(&buf, output.FormatJSON, quotes[i]); err != nil {
			t.Fatalf("WriteQuote: %v", err)
		}
		var row map[string]any
		if err := json.Unmarshal(buf.Bytes(), &row); err != nil {
			t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
		}
		got, present := row["security_type"]
		if row["symbol"] != w.symbol || present != w.present || (present && got != w.value) {
			t.Fatalf("row %d: symbol=%v security_type=(%v, present=%v), want symbol=%s security_type=(%q, present=%v)\n%s",
				i, row["symbol"], got, present, w.symbol, w.value, w.present, buf.String())
		}
	}
}

// enum 밖 값도 해석 · 거름 없이 원문 그대로 — 분류(미분류 갈래)는 소비자 몫임(a128 D2 §7 · a169 Q9).
func TestA169UnknownSecurityTypeIsPassedThroughUnjudged(t *testing.T) {
	quotes := adaptStocks([]apiStockInfo{{Symbol: "123456", SecurityType: "SOMETHING_NEW"}})
	var buf bytes.Buffer
	if err := output.WriteQuote(&buf, output.FormatJSON, quotes[0]); err != nil {
		t.Fatalf("WriteQuote: %v", err)
	}
	if !strings.Contains(buf.String(), `"security_type": "SOMETHING_NEW"`) {
		t.Fatalf("an unknown broker security type was not passed through verbatim:\n%s", buf.String())
	}
}

// 값이 실린 행의 바이트 — 키 이름 · 위치 고정(소비자가 읽는 JSON 경로).
func TestA169TheStocksQuoteJSONWithASecurityTypeIsPinned(t *testing.T) {
	quotes := adaptStocks([]apiStockInfo{{
		Symbol: "069500", Name: "KODEX 200", Market: "KOSPI", Currency: "KRW", Status: "ACTIVE", SecurityType: "ETF",
	}})
	quote := quotes[0]
	quote.FetchedAt = time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC) // 벽시계 고정
	var buf bytes.Buffer
	if err := output.WriteQuote(&buf, output.FormatJSON, quote); err != nil {
		t.Fatalf("WriteQuote: %v", err)
	}
	if buf.String() != a169StocksETFQuoteGolden {
		t.Fatalf("the /stocks quote JSON with a security type moved:\n got: %q\nwant: %q", buf.String(), a169StocksETFQuoteGolden)
	}
}

// 값을 받지 않는 생산자(/prices)는 키가 없어야 함 — 출력 바이트 불변(WTS 갈래는 internal/output a163 golden 이 고정).
func TestA169ThePricesQuoteJSONCarriesNoSecurityType(t *testing.T) {
	quotes := adaptPrices([]apiPrice{{Symbol: "005930", LastPrice: "72000", Currency: "KRW"}})
	var buf bytes.Buffer
	if err := output.WriteQuote(&buf, output.FormatJSON, quotes[0]); err != nil {
		t.Fatalf("WriteQuote: %v", err)
	}
	if strings.Contains(buf.String(), "security_type") {
		t.Fatalf("the /prices quote JSON carries a security type key:\n%s", buf.String())
	}
}
