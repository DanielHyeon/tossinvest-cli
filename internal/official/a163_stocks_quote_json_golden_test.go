package official

// a163 tasks 2.0(b) — 공식 /stocks 어댑터(adaptStocks)의 Quote 를 `tossctl quote --json` 과 같은 writer 로 쓴 바이트 고정.
// adaptStocks 는 브로커 데이터 시각을 채우지 않으므로 편집 전과 같은 바이트 · source_timestamp 키 없음이어야 함(omitempty — design D2).
// 골든은 편집 전 HEAD 에서 이 시험이 낸 바이트임.

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
)

const a163StocksQuoteGolden = "{\n  \"symbol\": \"005930\",\n  \"name\": \"삼성전자\",\n  \"market_code\": \"KOSPI\",\n  \"currency\": \"KRW\",\n  \"status\": \"ACTIVE\",\n  \"fetched_at\": \"2026-10-22T00:00:00Z\"\n}\n"

func TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime(t *testing.T) {
	quotes := adaptStocks([]apiStockInfo{{
		Symbol: "005930", Name: "삼성전자", EnglishName: "SamsungElec", Market: "KOSPI", Currency: "KRW", Status: "ACTIVE",
	}})
	if len(quotes) != 1 {
		t.Fatalf("adaptStocks = %d quotes, want 1", len(quotes))
	}
	quote := quotes[0]
	quote.FetchedAt = time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC) // 벽시계 고정
	var buf bytes.Buffer
	if err := output.WriteQuote(&buf, output.FormatJSON, quote); err != nil {
		t.Fatalf("WriteQuote: %v", err)
	}
	if strings.Contains(buf.String(), "source_timestamp") {
		t.Fatalf("the /stocks quote JSON carries a broker data time key:\n%s", buf.String())
	}
	if buf.String() != a163StocksQuoteGolden {
		t.Fatalf("the /stocks quote JSON moved:\n got: %q\nwant: %q", buf.String(), a163StocksQuoteGolden)
	}
}
