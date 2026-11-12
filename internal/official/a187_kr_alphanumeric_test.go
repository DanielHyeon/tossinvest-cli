package official

import (
	"context"
	"testing"
)

// a187: 두 엄격 호가 리더는 분봉 리더와 같은 한국 종목 문법(ASCII 대문자·숫자 6자리)을 씀.
// 공식 명세 1.2.24 의 호가 symbol 설명 「숫자 또는 영문·숫자 조합; 예: 005930, 0101N0」을 따름.
// 예전 거절 시험의 00593A·0101N0 두 사례는 여기서 수락으로 뜻을 뒤집어 고정함.

func a187KRPrices(symbol string) string {
	return strictQuotePricesBody(strictQuotePriceRow(symbol, "71000", "KRW", `"2026-08-18T09:06:26.400+09:00"`))
}

func TestA187StrictQuoteReadersAcceptKRAlphanumeric(t *testing.T) {
	t.Parallel()
	// 영문자 위치를 앞·가운데·뒤·전부로 흩음(리뷰 r1 발견 3: 앞 네 자리를 숫자로 좁히는 변이를 잡기 위함).
	for _, symbol := range []string{"0101N0", "00593A", "0158S0", "18064K", "A12345", "ABCDEF"} {
		t.Run(symbol, func(t *testing.T) {
			t.Parallel()
			book := newStrictMinuteBodyHarness(t, strictQuoteKROrderbook())
			top, err := book.client.StrictOrderbookTop(context.Background(), "KR", symbol)
			if err != nil {
				t.Fatalf("StrictOrderbookTop(KR, %s): %v", symbol, err)
			}
			if top.Symbol != symbol || top.Currency != "KRW" || book.hitCount() != 1 || book.lastQuery() != "symbol="+symbol {
				t.Fatalf("orderbook symbol=%q currency=%q hits=%d query=%q", top.Symbol, top.Currency, book.hitCount(), book.lastQuery())
			}
			prices := newStrictMinuteBodyHarness(t, a187KRPrices(symbol))
			last, err := prices.client.StrictLastPrice(context.Background(), "KR", symbol)
			if err != nil {
				t.Fatalf("StrictLastPrice(KR, %s): %v", symbol, err)
			}
			if last.Symbol != symbol || last.Currency != "KRW" || prices.hitCount() != 1 || prices.lastQuery() != "symbols="+symbol {
				t.Fatalf("prices symbol=%q currency=%q hits=%d query=%q", last.Symbol, last.Currency, prices.hitCount(), prices.lastQuery())
			}
		})
	}
}

// 영숫자 문법 밖의 KR 모양은 계속 요청 0회로 종목 사유와 함께 거절함. 대문자로 바꿔 주지 않음.
func TestA187StrictQuoteReadersStillRefuseOtherKRForms(t *testing.T) {
	t.Parallel()
	// 개행 두 모양은 여러 줄 모드 정규식(리뷰 r1 발견 3)을 잡기 위함.
	for _, symbol := range []string{"0101n0", " 0101N0", "0101N0 ", "0101N", "0101N00", "0101-0", "가101N0", "0101.0", "", "0101N0\n", "\n0101N0"} {
		t.Run(symbol, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, strictQuoteKROrderbook())
			if _, err := harness.client.StrictOrderbookTop(context.Background(), "KR", symbol); err == nil {
				t.Fatal("StrictOrderbookTop accepted")
			} else if refusal := strictQuoteRefusal(t, err); refusal.Reason != StrictQuoteReasonSymbol {
				t.Fatalf("StrictOrderbookTop reason = %q", refusal.Reason)
			}
			if _, err := harness.client.StrictLastPrice(context.Background(), "KR", symbol); err == nil {
				t.Fatal("StrictLastPrice accepted")
			} else if refusal := strictQuoteRefusal(t, err); refusal.Reason != StrictQuoteReasonSymbol {
				t.Fatalf("StrictLastPrice reason = %q", refusal.Reason)
			}
			if harness.hitCount() != 0 {
				t.Fatalf("requests = %d", harness.hitCount())
			}
		})
	}
}

// 숫자 KR 과 US 의 query·종목·통화는 이 change 전과 같음.
func TestA187NumericKRAndUSUnchanged(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		market, symbol, currency, book, prices string
	}{
		{"KR", "005930", "KRW", strictQuoteKROrderbook(), a187KRPrices("005930")},
		{"US", "AAPL", "USD", strictQuoteUSOrderbook(), strictQuoteUSPrices()},
	} {
		book := newStrictMinuteBodyHarness(t, c.book)
		top, err := book.client.StrictOrderbookTop(context.Background(), c.market, c.symbol)
		if err != nil || top.Symbol != c.symbol || top.Currency != c.currency || book.lastQuery() != "symbol="+c.symbol || book.hitCount() != 1 {
			t.Fatalf("%s orderbook: %+v %v query=%q", c.symbol, top, err, book.lastQuery())
		}
		prices := newStrictMinuteBodyHarness(t, c.prices)
		last, err := prices.client.StrictLastPrice(context.Background(), c.market, c.symbol)
		if err != nil || last.Symbol != c.symbol || last.Currency != c.currency || prices.lastQuery() != "symbols="+c.symbol || prices.hitCount() != 1 {
			t.Fatalf("%s prices: %+v %v query=%q", c.symbol, last, err, prices.lastQuery())
		}
	}
}

// KR 현재가도 응답 행의 종목 메아리가 요청과 다르면 거절함(리뷰 r1 발견 3: 기존 메아리 시험은 US 뿐이었음).
func TestA187StrictLastPriceRefusesKREchoMismatch(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t, a187KRPrices("0101N0"))
	if _, err := harness.client.StrictLastPrice(context.Background(), "KR", "0158S0"); err == nil {
		t.Fatal("StrictLastPrice accepted a row answering for another KR symbol")
	} else if refusal := strictQuoteRefusal(t, err); refusal.Reason != StrictQuoteReasonSymbolEcho {
		t.Fatalf("reason = %q, want %q", refusal.Reason, StrictQuoteReasonSymbolEcho)
	}
	if harness.hitCount() != 1 {
		t.Fatalf("requests = %d", harness.hitCount())
	}
}

// 분봉 리더도 같은 공용 문법으로 영숫자 KR 을 받고 종목마다 요청은 정확히 1회임(gstack 발견 2: 기존 분봉 수락 시험은
// 0101N0 하나에 요청 수 단언이 없었음). 래퍼가 공용 검사와 갈라지면 여기서 드러남.
func TestA187StrictMinuteCandlesAcceptKRAlphanumericOnce(t *testing.T) {
	t.Parallel()
	for _, symbol := range []string{"0158S0", "A12345", "18064K"} {
		t.Run(symbol, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, strictMinuteBody(`null`,
				strictMinuteCandleJSON("2026-08-14T09:31:00.000+09:00", "71000", "71100", "70900", "71050", "12", "KRW")))
			page, err := harness.client.StrictMinuteCandles(context.Background(), "KR", symbol, 200, "")
			if err != nil {
				t.Fatalf("StrictMinuteCandles(KR, %s): %v", symbol, err)
			}
			want := "adjusted=false&count=200&interval=1m&symbol=" + symbol
			if page.Symbol != symbol || harness.hitCount() != 1 || harness.lastQuery() != want {
				t.Fatalf("symbol=%q hits=%d query=%q", page.Symbol, harness.hitCount(), harness.lastQuery())
			}
		})
	}
}

// 분봉 리더도 영숫자 문법 밖의 KR 모양(소문자·개행)은 요청 0회로 거절함 — 래퍼가 자기 정규식을 다시 들이면 드러남.
func TestA187StrictMinuteCandlesStillRefuseOtherKRForms(t *testing.T) {
	t.Parallel()
	for _, symbol := range []string{"0158s0", "0101N0\n", "\n0101N0", "a12345"} {
		t.Run(symbol, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, strictMinuteBody(`null`))
			if _, err := harness.client.StrictMinuteCandles(context.Background(), "KR", symbol, 200, ""); err == nil {
				t.Fatal("StrictMinuteCandles accepted")
			} else if refusal := strictMinuteRefusal(t, err); refusal.Reason != StrictReasonSymbol {
				t.Fatalf("reason = %q", refusal.Reason)
			}
			if harness.hitCount() != 0 {
				t.Fatalf("requests = %d", harness.hitCount())
			}
		})
	}
}
