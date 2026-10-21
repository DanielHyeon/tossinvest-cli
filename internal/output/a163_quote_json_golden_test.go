package output

// a163 tasks 2.0 — `tossctl quote --json` 의 WTS(upstream) 갈래 바이트 고정(freeze 리뷰 P1-1, 안전 3).
//
// a163 은 domain.Quote 에 브로커 데이터 시각 필드를 더함. 그 필드를 채우지 않는 생산자(WTS GetQuote — internal/client/quote.go 의
// quote := domain.Quote{…})의 JSON 바이트가 편집 전과 같아야 하고 새 키가 없어야 함. 골든은 편집 전 HEAD 에서 이 시험이 낸 바이트임.
// FetchedAt 은 벽시계라 고정 시각으로 덮어 비교함. 공식 /stocks 어댑터 갈래는 internal/official 의 같은 이름 시험이 고정함.

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/client"
)

const a163WTSQuoteGolden = "{\n  \"product_code\": \"A005930\",\n  \"symbol\": \"005930\",\n  \"name\": \"삼성전자\",\n  \"market_code\": \"KSP\",\n  \"market\": \"코스피\",\n  \"currency\": \"KRW\",\n  \"reference_price\": 187900,\n  \"last\": 193900,\n  \"change\": 6000,\n  \"change_rate\": 0.031931878658861094,\n  \"volume\": 27306483,\n  \"open\": 189000,\n  \"high\": 195000,\n  \"low\": 186000,\n  \"high_52w\": 200000,\n  \"low_52w\": 56800,\n  \"market_cap\": 1150000000000000,\n  \"trading_value\": 5200000000000,\n  \"trading_strength\": 102.34,\n  \"prev_volume\": 25000000,\n  \"upper_limit\": 244000,\n  \"lower_limit\": 131600,\n  \"status\": \"N\",\n  \"fetched_at\": \"2026-10-22T00:00:00Z\"\n}\n"

func TestA163TheWTSQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime(t *testing.T) {
	fixtures := filepath.Join("..", "..", "fixtures", "responses", "public")
	routes := map[string]string{
		"/api/v2/stock-infos/A005930":            "stock-info.json",
		"/api/v1/stock-detail/ui/A005930/common": "stock-detail-common.json",
		"/api/v1/product/stock-prices":           "stock-price.json",
		"/api/v3/stock-prices/details":           "stock-price-details.json",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(fixtures, name))
	}))
	defer srv.Close()

	quote, err := client.New(client.Config{HTTPClient: srv.Client(), InfoBaseURL: srv.URL}).GetQuote(context.Background(), "005930")
	if err != nil {
		t.Fatalf("GetQuote: %v", err)
	}
	quote.FetchedAt = time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC) // 벽시계 고정
	var buf bytes.Buffer
	if err := WriteQuote(&buf, FormatJSON, quote); err != nil {
		t.Fatalf("WriteQuote: %v", err)
	}
	if strings.Contains(buf.String(), "source_timestamp") {
		t.Fatalf("the WTS quote JSON carries a broker data time key:\n%s", buf.String())
	}
	if buf.String() != a163WTSQuoteGolden {
		t.Fatalf("the WTS quote JSON moved (upstream output contract, safety 3):\n got: %q\nwant: %q", buf.String(), a163WTSQuoteGolden)
	}
}
