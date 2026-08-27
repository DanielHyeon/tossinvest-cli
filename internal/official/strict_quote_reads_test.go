package official

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// ---- 시험용 본문 조립기 ----
//
// 실측 영수증(receipt-2026-08-16-run4.json)이 보여 준 모양 그대로 만든다.
// 호가는 result가 객체 하나이고, 현재가는 result가 줄의 배열이다.

func strictQuoteLevel(price, volume string) string {
	return `{"price":"` + price + `","volume":"` + volume + `"}`
}

func strictQuoteOrderbookBody(timestamp, currency string, asks, bids []string) string {
	return `{"result":{"timestamp":` + timestamp + `,"currency":"` + currency +
		`","asks":[` + strings.Join(asks, ",") + `],"bids":[` + strings.Join(bids, ",") + `]}}`
}

func strictQuoteUSOrderbook() string {
	return strictQuoteOrderbookBody(`"2026-08-18T03:29:14.120+09:00"`, "USD",
		[]string{strictQuoteLevel("231.7000", "160")}, []string{strictQuoteLevel("231.6500", "40")})
}

// strictQuoteKROrderbook은 2026-08-18 KR 프로브가 본 모양이다. 매도는 낮은 값이 앞,
// 매수는 높은 값이 앞이므로 양쪽 모두 0번이 맨 위 한 줄이다.
func strictQuoteKROrderbook() string {
	asks := []string{strictQuoteLevel("284000", "10"), strictQuoteLevel("284500", "20"),
		strictQuoteLevel("285000", "30"), strictQuoteLevel("285500", "40"),
		strictQuoteLevel("286000", "50"), strictQuoteLevel("286500", "60"),
		strictQuoteLevel("287000", "70"), strictQuoteLevel("287500", "80"),
		strictQuoteLevel("288000", "90"), strictQuoteLevel("288500", "100")}
	bids := []string{strictQuoteLevel("283500", "11"), strictQuoteLevel("283000", "21"),
		strictQuoteLevel("282500", "31"), strictQuoteLevel("282000", "41"),
		strictQuoteLevel("281500", "51"), strictQuoteLevel("281000", "61"),
		strictQuoteLevel("280500", "71"), strictQuoteLevel("280000", "81"),
		strictQuoteLevel("279500", "91"), strictQuoteLevel("279000", "101")}
	return strictQuoteOrderbookBody(`"2026-08-18T09:06:26.500+09:00"`, "KRW", asks, bids)
}

func strictQuotePriceRow(symbol, last, currency, timestamp string) string {
	return `{"symbol":"` + symbol + `","lastPrice":"` + last + `","currency":"` + currency +
		`","timestamp":` + timestamp + `}`
}

func strictQuotePricesBody(rows ...string) string {
	return `{"result":[` + strings.Join(rows, ",") + `]}`
}

func strictQuoteUSPrices() string {
	return strictQuotePricesBody(strictQuotePriceRow("AAPL", "231.6800", "USD", `"2026-08-18T03:29:14.100+09:00"`))
}

func strictQuoteRefusal(t *testing.T, err error) *StrictQuoteError {
	t.Helper()
	var refusal *StrictQuoteError
	if !errors.As(err, &refusal) {
		t.Fatalf("error %v is not a StrictQuoteError", err)
	}
	return refusal
}

// ---- 요청을 보내기 전에 거절해야 하는 것들 ----

func TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name           string
		market, symbol string
	}{
		{"lower case market", "us", "AAPL"},
		{"unknown market", "JP", "AAPL"},
		{"empty market", "", "AAPL"},
		{"kr symbol with letters", "KR", "00593A"},
		{"kr symbol too short", "KR", "00593"},
		{"us symbol lower case", "US", "aapl"},
		{"us symbol too long", "US", "ABCDEFGHIJK"},
		{"empty symbol", "US", ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, strictQuoteUSOrderbook())
			if _, err := harness.client.StrictOrderbookTop(context.Background(), testCase.market, testCase.symbol); err == nil {
				t.Fatal("StrictOrderbookTop accepted an argument it must refuse")
			} else {
				strictQuoteRefusal(t, err)
			}
			if _, err := harness.client.StrictLastPrice(context.Background(), testCase.market, testCase.symbol); err == nil {
				t.Fatal("StrictLastPrice accepted an argument it must refuse")
			} else {
				strictQuoteRefusal(t, err)
			}
			if harness.hitCount() != 0 {
				t.Fatalf("the broker was called %d times for an argument the reader must refuse", harness.hitCount())
			}
		})
	}
}

func TestStrictQuoteReadersRefuseANilClient(t *testing.T) {
	t.Parallel()
	var client *Client
	if _, err := client.StrictOrderbookTop(context.Background(), "US", "AAPL"); err == nil {
		t.Fatal("a nil client returned an orderbook")
	} else {
		strictQuoteRefusal(t, err)
	}
	if _, err := client.StrictLastPrice(context.Background(), "US", "AAPL"); err == nil {
		t.Fatal("a nil client returned a price")
	} else {
		strictQuoteRefusal(t, err)
	}
}

// ---- 요청은 정확히 한 종목만 묻는다 ----

func TestStrictOrderbookTopAsksForExactlyOneSymbol(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t, strictQuoteKROrderbook())
	if _, err := harness.client.StrictOrderbookTop(context.Background(), "KR", "005930"); err != nil {
		t.Fatalf("StrictOrderbookTop: %v", err)
	}
	if harness.lastQuery() != "symbol=005930" {
		t.Fatalf("query = %q, want %q", harness.lastQuery(), "symbol=005930")
	}
}

func TestStrictLastPriceAsksForExactlyOneSymbol(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t, strictQuoteUSPrices())
	if _, err := harness.client.StrictLastPrice(context.Background(), "US", "AAPL"); err != nil {
		t.Fatalf("StrictLastPrice: %v", err)
	}
	if harness.lastQuery() != "symbols=AAPL" {
		t.Fatalf("query = %q, want %q", harness.lastQuery(), "symbols=AAPL")
	}
}

// ---- 받은 것을 그대로 돌려준다 ----

func TestStrictOrderbookTopReadsOnlyTheTopOfBook(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name           string
		body           string
		market, symbol string
		currency       string
		ask, bid       StrictQuoteLevel
		timestamp      string
	}{
		{
			name: "us one level per side", body: strictQuoteUSOrderbook(), market: "US", symbol: "AAPL",
			currency: "USD", ask: StrictQuoteLevel{Price: "231.7000", Volume: "160"},
			bid: StrictQuoteLevel{Price: "231.6500", Volume: "40"}, timestamp: "2026-08-18T03:29:14.120+09:00",
		},
		{
			name: "kr ten levels per side", body: strictQuoteKROrderbook(), market: "KR", symbol: "005930",
			currency: "KRW", ask: StrictQuoteLevel{Price: "284000", Volume: "10"},
			bid: StrictQuoteLevel{Price: "283500", Volume: "11"}, timestamp: "2026-08-18T09:06:26.500+09:00",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, testCase.body)
			top, err := harness.client.StrictOrderbookTop(context.Background(), testCase.market, testCase.symbol)
			if err != nil {
				t.Fatalf("StrictOrderbookTop: %v", err)
			}
			if top.Market != testCase.market || top.Symbol != testCase.symbol || top.Currency != testCase.currency {
				t.Fatalf("identity = %q/%q/%q", top.Market, top.Symbol, top.Currency)
			}
			if top.Ask != testCase.ask || top.Bid != testCase.bid {
				t.Fatalf("top of book = ask %+v bid %+v", top.Ask, top.Bid)
			}
			if top.SourceTimestamp != testCase.timestamp {
				t.Fatalf("source timestamp = %q, want %q", top.SourceTimestamp, testCase.timestamp)
			}
			if top.SourceInstant.IsZero() || top.SourceInstant.Format("2006-01-02T15:04:05.000-07:00") != testCase.timestamp {
				t.Fatalf("source instant = %v", top.SourceInstant)
			}
			if top.StatusCode != http.StatusOK || top.ReadAt.IsZero() {
				t.Fatalf("attempt = %d at %v", top.StatusCode, top.ReadAt)
			}
			sum := sha256.Sum256([]byte(testCase.body))
			if top.BodyDigest != "sha256:"+hex.EncodeToString(sum[:]) {
				t.Fatalf("body digest = %q", top.BodyDigest)
			}
		})
	}
}

func TestStrictLastPriceReadsTheOnlyRow(t *testing.T) {
	t.Parallel()
	body := strictQuoteUSPrices()
	harness := newStrictMinuteBodyHarness(t, body)
	last, err := harness.client.StrictLastPrice(context.Background(), "US", "AAPL")
	if err != nil {
		t.Fatalf("StrictLastPrice: %v", err)
	}
	if last.Market != "US" || last.Symbol != "AAPL" || last.Currency != "USD" || last.Last != "231.6800" {
		t.Fatalf("row = %+v", last)
	}
	if last.SourceTimestamp != "2026-08-18T03:29:14.100+09:00" || last.SourceInstant.IsZero() {
		t.Fatalf("instant = %q / %v", last.SourceTimestamp, last.SourceInstant)
	}
	if last.StatusCode != http.StatusOK || last.ReadAt.IsZero() {
		t.Fatalf("attempt = %d at %v", last.StatusCode, last.ReadAt)
	}
	sum := sha256.Sum256([]byte(body))
	if last.BodyDigest != "sha256:"+hex.EncodeToString(sum[:]) {
		t.Fatalf("body digest = %q", last.BodyDigest)
	}
}

// ---- 계약과 다른 본문은 고쳐 읽지 않고 거절한다 ----

func TestStrictOrderbookTopRefusesBodiesOutsideTheContract(t *testing.T) {
	t.Parallel()
	level := strictQuoteLevel("231.7000", "160")
	bid := strictQuoteLevel("231.6500", "40")
	instant := `"2026-08-18T03:29:14.120+09:00"`
	for _, testCase := range []struct {
		name string
		body string
	}{
		{"envelope duplicate key", `{"result":{"timestamp":` + instant + `,"currency":"USD","asks":[` + level + `],"bids":[` + bid + `]},"result":{}}`},
		{"result duplicate key", `{"result":{"timestamp":` + instant + `,"timestamp":` + instant + `,"currency":"USD","asks":[` + level + `],"bids":[` + bid + `]}}`},
		{"level duplicate key", strictQuoteOrderbookBody(instant, "USD",
			[]string{`{"price":"231.7000","price":"231.8000","volume":"160"}`}, []string{bid})},
		{"no result key", `{"payload":{"timestamp":` + instant + `,"currency":"USD","asks":[],"bids":[]}}`},
		{"result is not an object", `{"result":[1,2]}`},
		{"unknown result key", `{"result":{"timestamp":` + instant + `,"currency":"USD","asks":[` + level + `],"bids":[` + bid + `],"totalAskVolume":"160"}}`},
		{"missing result key", `{"result":{"currency":"USD","asks":[` + level + `],"bids":[` + bid + `]}}`},
		{"unknown level key", strictQuoteOrderbookBody(instant, "USD",
			[]string{`{"price":"231.7000","volume":"160","orders":"3"}`}, []string{bid})},
		{"missing level key", strictQuoteOrderbookBody(instant, "USD",
			[]string{`{"price":"231.7000"}`}, []string{bid})},
		{"bare number price", strictQuoteOrderbookBody(instant, "USD",
			[]string{`{"price":231.7,"volume":"160"}`}, []string{bid})},
		{"bare number volume", strictQuoteOrderbookBody(instant, "USD",
			[]string{`{"price":"231.7000","volume":160}`}, []string{bid})},
		{"level is not an object", strictQuoteOrderbookBody(instant, "USD", []string{`"231.7000"`}, []string{bid})},
		{"asks is not an array", `{"result":{"timestamp":` + instant + `,"currency":"USD","asks":{},"bids":[` + bid + `]}}`},
		{"empty asks", strictQuoteOrderbookBody(instant, "USD", nil, []string{bid})},
		{"empty bids", strictQuoteOrderbookBody(instant, "USD", []string{level}, nil)},
		{"null timestamp", strictQuoteOrderbookBody(`null`, "USD", []string{level}, []string{bid})},
		{"absent timestamp", `{"result":{"currency":"USD","asks":[` + level + `],"bids":[` + bid + `]}}`},
		{"zulu timestamp", strictQuoteOrderbookBody(`"2026-08-18T03:29:14Z"`, "USD", []string{level}, []string{bid})},
		{"offsetless timestamp", strictQuoteOrderbookBody(`"2026-08-18T03:29:14"`, "USD", []string{level}, []string{bid})},
		{"impossible timestamp", strictQuoteOrderbookBody(`"2026-02-30T03:29:14.000+09:00"`, "USD", []string{level}, []string{bid})},
		{"currency is not the market currency", strictQuoteOrderbookBody(instant, "KRW", []string{level}, []string{bid})},
		{"empty price", strictQuoteOrderbookBody(instant, "USD", []string{strictQuoteLevel("", "160")}, []string{bid})},
		{"over long price", strictQuoteOrderbookBody(instant, "USD",
			[]string{strictQuoteLevel(strings.Repeat("9", 31), "160")}, []string{bid})},
		{"trailing json value", strictQuoteUSOrderbook() + `{"result":{}}`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, testCase.body)
			if _, err := harness.client.StrictOrderbookTop(context.Background(), "US", "AAPL"); err == nil {
				t.Fatal("StrictOrderbookTop accepted a body outside the contract")
			} else {
				strictQuoteRefusal(t, err)
			}
		})
	}
}

func TestStrictLastPriceRefusesBodiesOutsideTheContract(t *testing.T) {
	t.Parallel()
	instant := `"2026-08-18T03:29:14.100+09:00"`
	row := strictQuotePriceRow("AAPL", "231.6800", "USD", instant)
	for _, testCase := range []struct {
		name string
		body string
	}{
		{"no rows", strictQuotePricesBody()},
		{"two rows", strictQuotePricesBody(row, strictQuotePriceRow("MSFT", "410.10", "USD", instant))},
		{"symbol does not echo", strictQuotePricesBody(strictQuotePriceRow("MSFT", "410.10", "USD", instant))},
		{"result is not an array", `{"result":` + row + `}`},
		{"no result key", `{"rows":[` + row + `]}`},
		{"row duplicate key", strictQuotePricesBody(`{"symbol":"AAPL","symbol":"AAPL","lastPrice":"231.6800","currency":"USD","timestamp":` + instant + `}`)},
		{"unknown row key", strictQuotePricesBody(`{"symbol":"AAPL","lastPrice":"231.6800","currency":"USD","timestamp":` + instant + `,"changeRate":"0.01"}`)},
		{"missing row key", strictQuotePricesBody(`{"symbol":"AAPL","lastPrice":"231.6800","currency":"USD"}`)},
		{"bare number last price", strictQuotePricesBody(`{"symbol":"AAPL","lastPrice":231.68,"currency":"USD","timestamp":` + instant + `}`)},
		{"row is not an object", strictQuotePricesBody(`"AAPL"`)},
		{"null timestamp", strictQuotePricesBody(strictQuotePriceRow("AAPL", "231.6800", "USD", `null`))},
		{"zulu timestamp", strictQuotePricesBody(strictQuotePriceRow("AAPL", "231.6800", "USD", `"2026-08-18T03:29:14Z"`))},
		{"currency is not the market currency", strictQuotePricesBody(strictQuotePriceRow("AAPL", "231.6800", "KRW", instant))},
		{"empty last price", strictQuotePricesBody(strictQuotePriceRow("AAPL", "", "USD", instant))},
		{"over long last price", strictQuotePricesBody(strictQuotePriceRow("AAPL", strings.Repeat("9", 31), "USD", instant))},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, testCase.body)
			if _, err := harness.client.StrictLastPrice(context.Background(), "US", "AAPL"); err == nil {
				t.Fatal("StrictLastPrice accepted a body outside the contract")
			} else {
				strictQuoteRefusal(t, err)
			}
		})
	}
}

func TestStrictQuoteReadersRefuseANonSuccessfulAttempt(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		read func(client *Client) error
	}{
		{"orderbook", func(client *Client) error {
			_, err := client.StrictOrderbookTop(context.Background(), "US", "AAPL")
			return err
		}},
		{"prices", func(client *Client) error {
			_, err := client.StrictLastPrice(context.Background(), "US", "AAPL")
			return err
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteHarness(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprint(w, `{"error":{"code":"rate-limited"}}`)
			})
			if err := testCase.read(harness.client); err == nil {
				t.Fatal("a non-2xx answer became evidence")
			}
		})
	}
}

// TestStrictQuoteRefusalsAreNeverCandleRefusals는 두 리더의 거절이 서로 섞이지 않게 한다.
// 호가가 거절됐는데 봉 거절로 보고되면, 무엇이 고장 났는지 읽는 쪽이 알 수 없다.
func TestStrictQuoteRefusalsAreNeverCandleRefusals(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t, strictQuoteUSOrderbook())
	_, err := harness.client.StrictOrderbookTop(context.Background(), "JP", "AAPL")
	if err == nil {
		t.Fatal("an unknown market was accepted")
	}
	var candle *StrictMinuteCandlesError
	if errors.As(err, &candle) {
		t.Fatalf("a quote refusal was reported as a candle refusal: %v", err)
	}
	if strictQuoteRefusal(t, err).Reason != StrictQuoteReasonMarket {
		t.Fatalf("reason = %q", strictQuoteRefusal(t, err).Reason)
	}
}
