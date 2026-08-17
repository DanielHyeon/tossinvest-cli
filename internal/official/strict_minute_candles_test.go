package official

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// ---- 테스트용 서버와 도우미들 ----

// strictMinuteHarness는 가짜 서버 하나와 그 서버를 보는 Client 하나를 묶어 둔다.
// 서버는 토큰 요청에는 늘 같은 토큰을 주고, 그 밖의 요청은 테스트가 준 함수에 넘긴다.
type strictMinuteHarness struct {
	client  *Client
	mu      sync.Mutex
	queries []string
	hits    int
}

func newStrictMinuteHarness(t *testing.T, respond func(w http.ResponseWriter, r *http.Request, hit int)) *strictMinuteHarness {
	t.Helper()
	harness := &strictMinuteHarness{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"token","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		harness.mu.Lock()
		harness.hits++
		hit := harness.hits
		harness.queries = append(harness.queries, r.URL.RawQuery)
		harness.mu.Unlock()
		respond(w, r, hit)
	}))
	t.Cleanup(server.Close)
	harness.client = New(Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token",
		WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	return harness
}

func newStrictMinuteBodyHarness(t *testing.T, body string) *strictMinuteHarness {
	t.Helper()
	return newStrictMinuteHarness(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		fmt.Fprint(w, body)
	})
}

func (h *strictMinuteHarness) hitCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hits
}

func (h *strictMinuteHarness) lastQuery() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.queries) == 0 {
		return ""
	}
	return h.queries[len(h.queries)-1]
}

func strictMinuteCandleJSON(timestamp, open, high, low, closePrice, volume, currency string) string {
	return `{"timestamp":"` + timestamp + `","openPrice":"` + open + `","highPrice":"` + high +
		`","lowPrice":"` + low + `","closePrice":"` + closePrice + `","volume":"` + volume +
		`","currency":"` + currency + `"}`
}

func strictMinuteUSCandle(timestamp, closePrice string) string {
	return strictMinuteCandleJSON(timestamp, "231.4350", "231.5", "231.4", closePrice, "1234", "USD")
}

func strictMinuteBody(cursor string, candles ...string) string {
	return `{"result":{"candles":[` + strings.Join(candles, ",") + `],"nextBefore":` + cursor + `}}`
}

func strictMinuteRefusal(t *testing.T, err error) *StrictMinuteCandlesError {
	t.Helper()
	var refusal *StrictMinuteCandlesError
	if !errors.As(err, &refusal) {
		t.Fatalf("error %v is not a StrictMinuteCandlesError", err)
	}
	return refusal
}

// ---- 요청을 보내기 전에 거절해야 하는 것들 ----

func TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name           string
		market, symbol string
		count          int
		before         string
	}{
		{"lower case market", "us", "AAPL", 200, ""},
		{"unknown market", "JP", "AAPL", 200, ""},
		{"kr symbol with letters", "KR", "00593A", 200, ""},
		{"kr symbol too short", "KR", "00593", 200, ""},
		{"us symbol lower case", "US", "aapl", 200, ""},
		{"us symbol too long", "US", "ABCDEFGHIJK", 200, ""},
		{"us symbol leading digit", "US", "1AAPL", 200, ""},
		{"empty symbol", "US", "", 200, ""},
		{"count zero", "US", "AAPL", 0, ""},
		{"count above the page cap", "US", "AAPL", 201, ""},
		{"count negative", "US", "AAPL", -1, ""},
		{"before with a zulu offset", "US", "AAPL", 200, "2026-08-14T15:07:00Z"},
		{"before without an offset", "US", "AAPL", 200, "2026-08-14T15:07:00"},
		{"before with four fractional digits", "US", "AAPL", 200, "2026-08-14T15:07:00.0000+09:00"},
		{"before with a named zone", "US", "AAPL", 200, "2026-08-14T15:07:00+09:00[Asia/Seoul]"},
		{"before that does not exist", "US", "AAPL", 200, "2026-02-30T15:07:00.000+09:00"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, strictMinuteBody(`null`))
			page, err := harness.client.StrictMinuteCandles(context.Background(),
				testCase.market, testCase.symbol, testCase.count, testCase.before)
			if err == nil {
				t.Fatalf("accepted %+v -> %+v", testCase, page)
			}
			strictMinuteRefusal(t, err)
			if harness.hitCount() != 0 {
				t.Fatalf("a refused argument still reached the broker: %d requests", harness.hitCount())
			}
		})
	}
}

// ---- 정상 경로 ----

func TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage(t *testing.T) {
	t.Parallel()
	body := strictMinuteBody(`"2026-08-14T22:31:00.000+09:00"`,
		strictMinuteUSCandle("2026-08-14T22:33:00.000+09:00", "231.44"),
		strictMinuteUSCandle("2026-08-14T22:32:00.000+09:00", "231.45"))
	harness := newStrictMinuteBodyHarness(t, body)
	page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200,
		"2026-08-15T05:00:00.000+09:00")
	if err != nil {
		t.Fatalf("StrictMinuteCandles: %v", err)
	}
	const wantQuery = "adjusted=false&before=2026-08-15T05%3A00%3A00.000%2B09%3A00&count=200&interval=1m&symbol=AAPL"
	if harness.lastQuery() != wantQuery {
		t.Fatalf("query = %q, want %q", harness.lastQuery(), wantQuery)
	}
	if page.Market != "US" || page.Symbol != "AAPL" || page.StatusCode != http.StatusOK {
		t.Fatalf("page identity = %+v", page)
	}
	if len(page.Candles) != 2 || page.Candles[0].Timestamp != "2026-08-14T22:33:00.000+09:00" ||
		page.Candles[0].Close != "231.44" || page.Candles[1].Currency != "USD" {
		t.Fatalf("candles = %+v", page.Candles)
	}
	if page.Terminal || page.NextBefore != "2026-08-14T22:31:00.000+09:00" {
		t.Fatalf("cursor = %q terminal=%v", page.NextBefore, page.Terminal)
	}
	digest := sha256.Sum256([]byte(body))
	if page.BodyDigest != "sha256:"+hex.EncodeToString(digest[:]) {
		t.Fatalf("body digest = %q", page.BodyDigest)
	}
	if page.ReadAt.IsZero() {
		t.Fatal("read-at instant is missing")
	}
}

func TestStrictMinuteCandlesAcceptsTheKoreanMarket(t *testing.T) {
	t.Parallel()
	body := strictMinuteBody(`null`,
		strictMinuteCandleJSON("2026-08-14T09:31:00.000+09:00", "71000", "71100", "70900", "71050", "12", "KRW"))
	harness := newStrictMinuteBodyHarness(t, body)
	page, err := harness.client.StrictMinuteCandles(context.Background(), "KR", "005930", 200, "")
	if err != nil {
		t.Fatalf("StrictMinuteCandles: %v", err)
	}
	if !page.Terminal || page.NextBefore != "" || len(page.Candles) != 1 || page.Candles[0].Currency != "KRW" {
		t.Fatalf("page = %+v", page)
	}
	if harness.lastQuery() != "adjusted=false&count=200&interval=1m&symbol=005930" {
		t.Fatalf("query = %q", harness.lastQuery())
	}
}

// 커서는 봉이 아니라 "다음 페이지의 상한"이다. 그래서 분 정렬 규칙은 커서에 적용하지
// 않는다. 실측 커서는 마지막 봉보다 1분 이른 분 경계였지만, 계약이 요구하는 것은
// "가장 오래된 봉보다 엄격히 이르다"뿐이다.
func TestStrictMinuteCandlesAcceptsASubMinuteCursor(t *testing.T) {
	t.Parallel()
	const cursor = "2026-08-14T22:32:59.999+09:00"
	harness := newStrictMinuteBodyHarness(t, strictMinuteBody(`"`+cursor+`"`,
		strictMinuteUSCandle("2026-08-14T22:33:00.000+09:00", "231.44")))
	page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
	if err != nil {
		t.Fatalf("a sub-minute cursor was refused: %v", err)
	}
	if page.Terminal || page.NextBefore != cursor {
		t.Fatalf("page = %+v", page)
	}
}

func TestStrictMinuteCandlesAcceptsAnEmptyPageThatStillCarriesACursor(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t, strictMinuteBody(`"2026-08-14T22:31:00.000+09:00"`))
	page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
	if err != nil {
		t.Fatalf("StrictMinuteCandles: %v", err)
	}
	if len(page.Candles) != 0 || page.Terminal || page.NextBefore != "2026-08-14T22:31:00.000+09:00" {
		t.Fatalf("page = %+v", page)
	}
}

// ---- 본문 계약 위반 ----

func TestStrictMinuteCandlesRefusesMalformedBodies(t *testing.T) {
	t.Parallel()
	twoNewestFirst := []string{
		strictMinuteUSCandle("2026-08-14T22:33:00.000+09:00", "231.44"),
		strictMinuteUSCandle("2026-08-14T22:32:00.000+09:00", "231.45"),
	}
	for _, testCase := range []struct{ name, body, want string }{
		{"duplicate key in the envelope", `{"result":{"candles":[],"nextBefore":null},"result":{"candles":[],"nextBefore":null}}`, StrictReasonBody},
		{"duplicate key in the result", `{"result":{"candles":[],"nextBefore":null,"candles":[]}}`, StrictReasonBody},
		{"duplicate key in a candle", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":"1","highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1","currency":"USD"}],"nextBefore":null}}`, StrictReasonBody},
		{"result absent", `{"data":{"candles":[],"nextBefore":null}}`, StrictReasonResult},
		{"result is not an object", `{"result":[]}`, StrictReasonResult},
		{"unknown result key", `{"result":{"candles":[],"nextBefore":null,"hasMore":false}}`, StrictReasonResult},
		{"candles absent", `{"result":{"nextBefore":null}}`, StrictReasonResult},
		{"candles is not an array", `{"result":{"candles":{},"nextBefore":null}}`, StrictReasonResult},
		{"candles is null", `{"result":{"candles":null,"nextBefore":null}}`, StrictReasonResult},
		{"unknown candle key", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":"1","highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1","currency":"USD","tradeCount":"3"}],"nextBefore":null}}`, StrictReasonCandle},
		{"missing candle key", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":"1","highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1"}],"nextBefore":null}}`, StrictReasonCandle},
		{"candle is not an object", `{"result":{"candles":["x"],"nextBefore":null}}`, StrictReasonCandle},
		{"bare number price", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":1,"highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1","currency":"USD"}],"nextBefore":null}}`, StrictReasonCandle},
		{"bare number volume", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":"1","highPrice":"1","lowPrice":"1","closePrice":"1","volume":1,"currency":"USD"}],"nextBefore":null}}`, StrictReasonCandle},
		{"null price", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":null,"highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1","currency":"USD"}],"nextBefore":null}}`, StrictReasonCandle},
		{"empty price string", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":"","highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1","currency":"USD"}],"nextBefore":null}}`, StrictReasonCandle},
		{"over long decimal", `{"result":{"candles":[{"timestamp":"2026-08-14T22:33:00.000+09:00","openPrice":"1234567890123456789012345678901","highPrice":"1","lowPrice":"1","closePrice":"1","volume":"1","currency":"USD"}],"nextBefore":null}}`, StrictReasonCandle},
		{"foreign currency", strictMinuteBody(`null`, strictMinuteCandleJSON("2026-08-14T22:33:00.000+09:00", "1", "1", "1", "1", "1", "KRW")), StrictReasonCandle},
		{"timestamp without an offset", strictMinuteBody(`null`, strictMinuteUSCandle("2026-08-14T22:33:00.000", "1")), StrictReasonCandle},
		{"timestamp with a zulu offset", strictMinuteBody(`null`, strictMinuteUSCandle("2026-08-14T13:33:00.000Z", "1")), StrictReasonCandle},
		{"timestamp that does not exist", strictMinuteBody(`null`, strictMinuteUSCandle("2026-02-30T22:33:00.000+09:00", "1")), StrictReasonCandle},
		{"ascending instants", strictMinuteBody(`null`, twoNewestFirst[1], twoNewestFirst[0]), StrictReasonOrder},
		{"duplicate instants", strictMinuteBody(`null`, twoNewestFirst[0], twoNewestFirst[0]), StrictReasonOrder},
		{"cursor absent", `{"result":{"candles":[]}}`, StrictReasonResult},
		// 열쇠는 둘인데 커서가 아닌 다른 이름이 들어온 경우. 열쇠 개수 규칙에 걸리지 않으므로
		// "없는 커서는 마지막 페이지가 아니다"라는 규칙 하나만으로 막아야 한다.
		{"cursor replaced by another key", `{"result":{"candles":[],"hasMore":false}}`, StrictReasonCursor},
		{"cursor is an empty string", strictMinuteBody(`""`), StrictReasonCursor},
		{"cursor is a number", strictMinuteBody(`17`), StrictReasonCursor},
		{"cursor is an object", strictMinuteBody(`{}`), StrictReasonCursor},
		{"cursor is an array", strictMinuteBody(`[]`), StrictReasonCursor},
		{"cursor is a boolean", strictMinuteBody(`false`), StrictReasonCursor},
		{"cursor breaks the timestamp grammar", strictMinuteBody(`"2026-08-14T22:31:00Z"`), StrictReasonCursor},
		{"cursor is not older than the oldest bar", strictMinuteBody(`"2026-08-14T22:32:00.000+09:00"`, twoNewestFirst...), StrictReasonCursor},
		{"cursor is newer than the oldest bar", strictMinuteBody(`"2026-08-14T22:34:00.000+09:00"`, twoNewestFirst...), StrictReasonCursor},
		{"trailing json value", strictMinuteBody(`null`) + `{}`, StrictReasonBody},
		{"body is not an object", `[]`, StrictReasonBody},
		{"empty body", ``, StrictReasonBody},
		// 봉투의 무시되는 형제 안에 있는 중복 열쇠. 이것은 전체를 훑는 검사만 잡을 수 있다.
		{"duplicate key inside an ignored sibling", `{"meta":{"x":1,"x":2},"result":{"candles":[],"nextBefore":null}}`, StrictReasonBody},
		{"seventeen levels of nesting", `{"meta":` + nestedJSONObject(17) + `,"result":{"candles":[],"nextBefore":null}}`, StrictReasonBody},
		{"instant with non-zero seconds", strictMinuteBody(`null`, strictMinuteUSCandle("2026-08-14T22:33:30.000+09:00", "1")), StrictReasonCandleNotOnMinute},
		{"instant with a fraction", strictMinuteBody(`null`, strictMinuteUSCandle("2026-08-14T22:33:00.500+09:00", "1")), StrictReasonCandleNotOnMinute},
		// 가격 문자열 안의 깨진 바이트. 표준 해독기는 이것을 U+FFFD로 바꿔 삼킨다.
		{"invalid utf-8 inside a price", strictMinuteBody(`null`,
			strictMinuteCandleJSON("2026-08-14T22:33:00.000+09:00", "231.4\xff350", "1", "1", "1", "1", "USD")), StrictReasonInvalidUTF8},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteBodyHarness(t, testCase.body)
			page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
			if err == nil {
				t.Fatalf("accepted %s -> %+v", testCase.name, page)
			}
			if got := strictMinuteRefusal(t, err); got.Reason != testCase.want {
				t.Fatalf("reason = %q, want %q (%v)", got.Reason, testCase.want, got)
			}
		})
	}
}

// nestedJSONObject는 중괄호가 depth개 겹친 값을 만든다.
func nestedJSONObject(depth int) string {
	value := "1"
	for index := 0; index < depth; index++ {
		value = `{"a":` + value + `}`
	}
	return value
}

func TestStrictMinuteCandlesAcceptsNestingUpToTheBound(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t,
		`{"meta":`+nestedJSONObject(16)+`,"result":{"candles":[],"nextBefore":null}}`)
	page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
	if err != nil || !page.Terminal {
		t.Fatalf("sixteen levels of nesting were refused: %v (%+v)", err, page)
	}
}

func TestStrictMinuteCandlesRefusesInvalidUTF8(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteHarness(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		_, _ = w.Write([]byte("{\"result\":{\"candles\":[],\"nextBefore\":\"\xff\xfe\"}}"))
	})
	if _, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, ""); err == nil {
		t.Fatal("accepted a body that is not valid UTF-8")
	} else {
		strictMinuteRefusal(t, err)
	}
}

func TestStrictMinuteCandlesRefusesMoreCandlesThanRequested(t *testing.T) {
	t.Parallel()
	body := strictMinuteBody(`null`,
		strictMinuteUSCandle("2026-08-14T22:33:00.000+09:00", "1"),
		strictMinuteUSCandle("2026-08-14T22:32:00.000+09:00", "1"),
		strictMinuteUSCandle("2026-08-14T22:31:00.000+09:00", "1"))
	harness := newStrictMinuteBodyHarness(t, body)
	if _, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 2, ""); err == nil {
		t.Fatal("accepted three candles for a count of two")
	} else {
		strictMinuteRefusal(t, err)
	}
	if _, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 3, ""); err != nil {
		t.Fatalf("refused three candles for a count of three: %v", err)
	}
}

func TestStrictMinuteCandlesRefusesInstantsNewerThanBefore(t *testing.T) {
	t.Parallel()
	body := strictMinuteBody(`null`, strictMinuteUSCandle("2026-08-14T22:33:00.000+09:00", "1"))
	harness := newStrictMinuteBodyHarness(t, body)
	if _, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200,
		"2026-08-14T22:32:00.000+09:00"); err == nil {
		t.Fatal("accepted a bar newer than the inclusive upper bound")
	} else {
		strictMinuteRefusal(t, err)
	}
	// 상한과 정확히 같은 봉은 받아들인다. 상한은 포함 상한이기 때문이다.
	if _, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200,
		"2026-08-14T22:33:00.000+09:00"); err != nil {
		t.Fatalf("refused the inclusive upper bound itself: %v", err)
	}
}

func TestStrictMinuteCandlesRefusesABodyAboveTwoMebibytes(t *testing.T) {
	t.Parallel()
	padding := strings.Repeat("a", 2<<20)
	harness := newStrictMinuteBodyHarness(t, `{"pad":"`+padding+`","result":{"candles":[],"nextBefore":null}}`)
	if _, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, ""); err == nil {
		t.Fatal("accepted a body above the cap")
	} else {
		strictMinuteRefusal(t, err)
	}
}

func TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteBodyHarness(t, `{"traceId":"abc","result":{"candles":[],"nextBefore":null}}`)
	page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
	if err != nil || !page.Terminal {
		t.Fatalf("page = %+v err = %v", page, err)
	}
}

// ---- 전송 계층과의 연결 ----

func TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver(t *testing.T) {
	t.Parallel()
	const secondBody = `{"result":{"candles":[],"nextBefore":null}}`
	harness := newStrictMinuteHarness(t, func(w http.ResponseWriter, _ *http.Request, hit int) {
		if hit == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"unauthorized"}`)
			return
		}
		fmt.Fprint(w, secondBody)
	})
	var outer []AttemptTrace
	ctx := WithAttemptObserver(context.Background(), func(trace AttemptTrace) { outer = append(outer, trace) })
	page, err := harness.client.StrictMinuteCandles(ctx, "US", "AAPL", 200, "")
	if err != nil {
		t.Fatalf("StrictMinuteCandles: %v", err)
	}
	if len(outer) != 2 || outer[0].StatusCode != http.StatusUnauthorized || outer[1].StatusCode != http.StatusOK {
		t.Fatalf("outer observer was shadowed: %+v", outer)
	}
	digest := sha256.Sum256([]byte(secondBody))
	if page.BodyDigest != "sha256:"+hex.EncodeToString(digest[:]) || page.StatusCode != http.StatusOK {
		t.Fatalf("page did not come from the last successful attempt: %+v", page)
	}
	if !page.ReadAt.Equal(outer[1].BodyReadComplete) {
		t.Fatalf("read-at %v is not the second attempt's body-read instant %v", page.ReadAt, outer[1].BodyReadComplete)
	}
}

func TestStrictMinuteCandlesPropagatesTransportClassification(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		status int
		want   error
	}{
		{"rate limited", http.StatusTooManyRequests, ErrRateLimited},
		{"server error", http.StatusInternalServerError, ErrServer},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			harness := newStrictMinuteHarness(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
				w.WriteHeader(testCase.status)
				fmt.Fprint(w, `{"message":"no"}`)
			})
			page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
			if !errors.Is(err, testCase.want) {
				t.Fatalf("err = %v, want %v", err, testCase.want)
			}
			var refusal *StrictMinuteCandlesError
			if errors.As(err, &refusal) {
				t.Fatalf("a transport failure was reported as a contract refusal: %v", refusal)
			}
			if len(page.Candles) != 0 || page.BodyDigest != "" {
				t.Fatalf("a failed read still produced a page: %+v", page)
			}
		})
	}
}

func TestStrictMinuteCandlesReportsTheRateBudgetOfTheSameResponse(t *testing.T) {
	t.Parallel()
	harness := newStrictMinuteHarness(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		w.Header().Set(HeaderRateLimit, "20")
		w.Header().Set(HeaderRateRemaining, "18")
		w.Header().Set(HeaderRateReset, "1")
		fmt.Fprint(w, `{"result":{"candles":[],"nextBefore":null}}`)
	})
	page, err := harness.client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, "")
	if err != nil {
		t.Fatalf("StrictMinuteCandles: %v", err)
	}
	if !page.Budget.Reported || page.Budget.Limit != 20 || page.Budget.Remaining != 18 || page.Budget.Exhausted() {
		t.Fatalf("budget = %+v", page.Budget)
	}
}

func TestStrictMinuteCandlesRefusesANilClient(t *testing.T) {
	t.Parallel()
	var client *Client
	if _, err := client.StrictMinuteCandles(context.Background(), "US", "AAPL", 200, ""); err == nil {
		t.Fatal("a nil client produced a page")
	}
}

// ---- 정적 가드 ----

func TestStrictMinutePageLiteralLivesOnlyInItsOwnFile(t *testing.T) {
	t.Parallel()
	dir := strictMinutePackageDir(t)
	entries, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	scanned := 0
	for _, path := range entries {
		name := filepath.Base(path)
		if strings.HasSuffix(name, "_test.go") || name == "strict_minute_candles.go" {
			continue
		}
		scanned++
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(source), "StrictMinutePage{") {
			t.Fatalf("%s builds a StrictMinutePage; only the reader may", name)
		}
	}
	// 훑은 파일이 0개이면 이 시험은 아무것도 증명하지 않는다.
	if scanned < 10 {
		t.Fatalf("only %d production files were scanned", scanned)
	}
}

func TestStrictMinuteCandlesFilesCarryNoSeamIdentifier(t *testing.T) {
	t.Parallel()
	// 이 needle은 일부러 두 조각으로 나눠 이어 붙인다. 통째로 적으면 이 파일 자신이
	// 금지된 바이트열을 갖게 되기 때문이다.
	needle := "a112" + "MBUS"
	dir := strictMinutePackageDir(t)
	for _, name := range []string{"strict_minute_candles.go", "strict_minute_candles_test.go"} {
		source, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(source), needle) {
			t.Fatalf("%s references the measurement seam", name)
		}
	}
}

func strictMinutePackageDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	return filepath.Dir(file)
}
