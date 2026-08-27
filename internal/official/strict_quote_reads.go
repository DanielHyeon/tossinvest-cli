package official

// strict_quote_reads.go는 호가창 맨 위 한 줄과 현재가를 "받은 바이트 그대로" 엄격하게 읽는다.
//
// 왜 따로 만들었나
//
// 기존 Orderbook/Prices는 화면과 도메인을 위한 길이다. 그 길은 브로커가 보낸 시각(timestamp)과
// 통화를 **버리고**, 소수 문자열을 parseDecimal로 float64에 담는다. parseDecimal은 오류를
// 조용히 0으로 바꾸므로, 그 0이 브로커가 보낸 0인지 우리가 만든 0인지 구별할 수 없다.
// 증거로 쓰려면 그 셋(없음·null·0)이 서로 다른 사건이어야 한다. 그래서 이 파일은 응답 본문의
// 정확한 바이트를 직접 읽고, 조금이라도 계약과 다르면 고쳐 읽지 않고 거절한다(fail closed).
//
// 무엇을 읽나 (a112 결정 33)
//
// 증거의 이름이 official_quote_l1이고 L1은 한 줄이라는 뜻이다. 이 리더는 asks[0]과 bids[0]만
// 읽는다. 사다리의 깊이는 시장마다 다르므로(2026-08-18 실측: KR 10줄, US 1줄) 깊이는 응답
// 데이터이지 계약 상수가 아니다. 세지도, 상한을 두지도, 합계를 만들지도 않는다.
//
// 무엇을 건드리지 않나
//
// 토큰·재시도·전송은 평소 생산 경로(c.get → send → doRequest)를 그대로 쓴다. 모양 검사기
// (strictMinute*)도 그대로 **부른다** — 고치지 않는다. 다만 거절은 이 파일의 오류 타입으로
// 낸다. 호가가 거절됐는데 봉 거절로 보고되면 무엇이 고장 났는지 읽는 쪽이 알 수 없다.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"
	"unicode/utf8"
)

// 요청 경로. 요청과 rate budget 조회가 같은 글자를 써야 예산이 조용히 빗나가지 않는다.
const (
	PathOrderbook = "/api/v1/orderbook"
	PathPrices    = "/api/v1/prices"
)

// 계약이 고정한 열쇠들. 더도 덜도 안 된다.
var (
	strictQuoteOrderbookKeys = []string{"timestamp", "currency", "asks", "bids"}
	strictQuoteLevelKeys     = []string{"price", "volume"}
	strictQuotePriceRowKeys  = []string{"symbol", "lastPrice", "currency", "timestamp"}
)

// StrictQuoteLevel은 호가 한 줄이다. 값은 브로커가 보낸 소수 문자열 그대로다.
type StrictQuoteLevel struct {
	Price  string
	Volume string
}

// StrictTopOfBook은 호가창 맨 위 한 줄의 읽기 결과다.
type StrictTopOfBook struct {
	Market   string
	Symbol   string
	Currency string
	// Ask는 asks[0], Bid는 bids[0]이다. 그 아래 줄은 읽지 않는다(결정 33).
	Ask StrictQuoteLevel
	Bid StrictQuoteLevel
	// SourceTimestamp는 브로커가 보낸 글자 그대로의 시각이다.
	SourceTimestamp string
	// SourceInstant는 그 글자를 해독한 시각이다. 우리 시계가 아니라 브로커의 시계다.
	SourceInstant time.Time
	// ReadAt은 이 본문을 다 읽은 순간이다(응답에 묶인 우리 시각).
	ReadAt     time.Time
	StatusCode int
	BodyDigest string
	Budget     RateBudget
}

// StrictLastPrice는 현재가 한 줄의 읽기 결과다.
type StrictLastPrice struct {
	Market          string
	Symbol          string
	Currency        string
	Last            string
	SourceTimestamp string
	SourceInstant   time.Time
	ReadAt          time.Time
	StatusCode      int
	BodyDigest      string
	Budget          RateBudget
}

// StrictQuoteError는 "계약이 어긋나서 거절했다"는 뜻이다.
type StrictQuoteError struct {
	Reason string
	Detail string
}

func (e *StrictQuoteError) Error() string {
	return "official strict quote: " + e.Reason + ": " + e.Detail
}

// 거절 사유 이름표.
const (
	StrictQuoteReasonClientMissing   = "CLIENT_MISSING"
	StrictQuoteReasonMarket          = "MARKET_INVALID"
	StrictQuoteReasonSymbol          = "SYMBOL_INVALID"
	StrictQuoteReasonNoAttempt       = "NO_SUCCESSFUL_ATTEMPT"
	StrictQuoteReasonBodyTooLarge    = "BODY_TOO_LARGE"
	StrictQuoteReasonInvalidUTF8     = "BODY_NOT_UTF8"
	StrictQuoteReasonBody            = "BODY_INVALID"
	StrictQuoteReasonResult          = "RESULT_INVALID"
	StrictQuoteReasonLevel           = "LEVEL_INVALID"
	StrictQuoteReasonEmptySide       = "EMPTY_SIDE"
	StrictQuoteReasonNoSourceInstant = "NO_SOURCE_INSTANT"
	StrictQuoteReasonCurrency        = "CURRENCY_MISMATCH"
	StrictQuoteReasonRow             = "ROW_INVALID"
	StrictQuoteReasonRowCount        = "ROW_COUNT_INVALID"
	StrictQuoteReasonSymbolEcho      = "SYMBOL_ECHO_MISMATCH"
	StrictQuoteReasonDecimal         = "DECIMAL_INVALID"
)

func strictQuoteRefuse(reason, detail string) error {
	return &StrictQuoteError{Reason: reason, Detail: detail}
}

// strictQuoteAdopt는 모양 검사기가 낸 거절을 호가 거절로 바꿔 든다.
// 문법의 권위는 한 곳(strictMinute*)에 두고, 오류 타입만 여기 것으로 만든다.
func strictQuoteAdopt(reason string, err error) error {
	var candle *StrictMinuteCandlesError
	if errors.As(err, &candle) {
		return strictQuoteRefuse(reason, candle.Detail)
	}
	return strictQuoteRefuse(reason, err.Error())
}

// StrictOrderbookTop은 호가창 맨 위 한 줄을 엄격하게 읽는다.
//
// 요청을 보내기 전에 인자 문법을 모두 확인한다. 문법이 틀린 요청은 브로커에게 보내지
// 않는다(공유 쿼터를 낭비하지 않는다).
func (c *Client) StrictOrderbookTop(ctx context.Context, market, symbol string) (StrictTopOfBook, error) {
	if c == nil {
		return StrictTopOfBook{}, strictQuoteRefuse(StrictQuoteReasonClientMissing, "client is required")
	}
	currency, err := strictQuoteIdentity(market, symbol)
	if err != nil {
		return StrictTopOfBook{}, err
	}
	used, err := c.strictQuoteFetch(ctx, PathOrderbook, url.Values{"symbol": {symbol}})
	if err != nil {
		return StrictTopOfBook{}, err
	}
	ask, bid, stamp, instant, err := strictQuoteDecodeOrderbook(used.Body, currency)
	if err != nil {
		return StrictTopOfBook{}, err
	}
	digest := sha256.Sum256(used.Body)
	return StrictTopOfBook{
		Market: market, Symbol: symbol, Currency: currency,
		Ask: ask, Bid: bid, SourceTimestamp: stamp, SourceInstant: instant,
		ReadAt: used.BodyReadComplete, StatusCode: used.StatusCode,
		BodyDigest: "sha256:" + hex.EncodeToString(digest[:]),
		Budget:     c.RateBudget(PathOrderbook),
	}, nil
}

// StrictLastPrice는 한 종목의 현재가를 엄격하게 읽는다.
//
// 정확히 한 종목만 묻고, 정확히 한 줄만 받는다. Client.Prices는 모르는 종목에 빈 목록과
// nil 오류를 돌려주므로(그 침묵이 이 리더가 따로 있는 이유다) 여기서는 줄 수와 종목 메아리를
// 직접 확인한다.
func (c *Client) StrictLastPrice(ctx context.Context, market, symbol string) (StrictLastPrice, error) {
	if c == nil {
		return StrictLastPrice{}, strictQuoteRefuse(StrictQuoteReasonClientMissing, "client is required")
	}
	currency, err := strictQuoteIdentity(market, symbol)
	if err != nil {
		return StrictLastPrice{}, err
	}
	used, err := c.strictQuoteFetch(ctx, PathPrices, url.Values{"symbols": {symbol}})
	if err != nil {
		return StrictLastPrice{}, err
	}
	last, stamp, instant, err := strictQuoteDecodePrices(used.Body, symbol, currency)
	if err != nil {
		return StrictLastPrice{}, err
	}
	digest := sha256.Sum256(used.Body)
	return StrictLastPrice{
		Market: market, Symbol: symbol, Currency: currency, Last: last,
		SourceTimestamp: stamp, SourceInstant: instant,
		ReadAt: used.BodyReadComplete, StatusCode: used.StatusCode,
		BodyDigest: "sha256:" + hex.EncodeToString(digest[:]),
		Budget:     c.RateBudget(PathPrices),
	}, nil
}

// strictQuoteIdentity는 시장과 종목 문법을 한 번에 본다. 문법의 권위는 봉 리더와 같다.
func strictQuoteIdentity(market, symbol string) (string, error) {
	currency, err := strictMinuteMarketCurrency(market)
	if err != nil {
		return "", strictQuoteAdopt(StrictQuoteReasonMarket, err)
	}
	if err := strictMinuteCheckSymbol(market, symbol); err != nil {
		return "", strictQuoteAdopt(StrictQuoteReasonSymbol, err)
	}
	return currency, nil
}

// strictQuoteFetch는 GET 한 번을 보내고 그 시도의 원본 바이트를 받아 온다.
// 이미 ctx에 관측자가 있으면 그것도 계속 부른다(덮어쓰지 않는다).
func (c *Client) strictQuoteFetch(ctx context.Context, path string, query url.Values) (AttemptTrace, error) {
	outer, _ := ctx.Value(attemptObserverKey{}).(AttemptObserver)
	var attempts []AttemptTrace
	traced := WithAttemptObserver(ctx, func(trace AttemptTrace) {
		attempts = append(attempts, trace)
		if outer != nil {
			outer(trace)
		}
	})
	if err := c.get(traced, path, query, nil); err != nil {
		return AttemptTrace{}, err
	}
	used, err := strictMinuteFinalAttempt(attempts)
	if err != nil {
		return AttemptTrace{}, strictQuoteAdopt(StrictQuoteReasonNoAttempt, err)
	}
	return used, nil
}

// strictQuoteResult는 봉투를 열고 result 하나를 원본 바이트로 꺼낸다.
func strictQuoteResult(body []byte) (json.RawMessage, error) {
	if len(body) > strictMinuteMaxBody {
		return nil, strictQuoteRefuse(StrictQuoteReasonBodyTooLarge,
			"body is "+strconv.Itoa(len(body))+" bytes, above the cap of "+strconv.Itoa(strictMinuteMaxBody))
	}
	// 표준 해독기는 깨진 바이트를 U+FFFD로 조용히 바꾼다. 그러면 브로커가 보내지 않은
	// 값이 증거가 되므로, 바이트를 보기 전에 먼저 거절한다.
	if !utf8.Valid(body) {
		return nil, strictQuoteRefuse(StrictQuoteReasonInvalidUTF8, "body is not valid UTF-8")
	}
	if err := strictMinuteCheckJSON(body); err != nil {
		return nil, strictQuoteRefuse(StrictQuoteReasonBody, err.Error())
	}
	envelope, err := strictMinuteObject(body)
	if err != nil {
		return nil, strictQuoteRefuse(StrictQuoteReasonBody, "envelope: "+err.Error())
	}
	// 봉투의 다른 열쇠는 무시한다. result만이 계약이다.
	resultRaw, found := envelope["result"]
	if !found {
		return nil, strictQuoteRefuse(StrictQuoteReasonResult, "envelope has no 'result' key")
	}
	return resultRaw, nil
}

// strictQuoteFields는 객체 하나를 열고 계약이 정한 열쇠가 정확히 그것뿐인지 본다.
func strictQuoteFields(raw []byte, keys []string, reason, what string) (map[string]json.RawMessage, error) {
	fields, err := strictMinuteObject(raw)
	if err != nil {
		return nil, strictQuoteRefuse(reason, what+": "+err.Error())
	}
	if len(fields) != len(keys) {
		return nil, strictQuoteRefuse(reason,
			what+" carries "+strconv.Itoa(len(fields))+" keys, not the "+
				strconv.Itoa(len(keys))+" the contract fixes")
	}
	for _, key := range keys {
		if _, found := fields[key]; !found {
			return nil, strictQuoteRefuse(reason, what+" has no "+strconv.Quote(key)+" key")
		}
	}
	return fields, nil
}

// strictQuoteText는 값 하나를 JSON 문자열로 꺼낸다. 숫자·null·객체는 전부 거절이다.
func strictQuoteText(raw json.RawMessage, reason, what string) (string, error) {
	text, err := strictMinuteString(raw)
	if err != nil {
		return "", strictQuoteRefuse(reason, what+": "+err.Error())
	}
	return text, nil
}

// strictQuoteDecimal은 소수 문자열의 길이만 본다. 숫자 문법의 유일한 권위는
// strategyevidence이므로 여기서 값을 해석하지 않는다.
func strictQuoteDecimal(raw json.RawMessage, what string) (string, error) {
	text, err := strictQuoteText(raw, StrictQuoteReasonDecimal, what)
	if err != nil {
		return "", err
	}
	if len(text) > strictMinuteMaxDecimal {
		return "", strictQuoteRefuse(StrictQuoteReasonDecimal,
			what+" is longer than "+strconv.Itoa(strictMinuteMaxDecimal)+" bytes")
	}
	return text, nil
}

// strictQuoteInstant는 브로커 시각을 요구한다.
//
// 우리 시계로 대신하지 않는다(결정 35). validateQuote의 신선도 거부는
// evaluated − source_observed_at을 재는데, 그 값이 우리 읽은 시각이면 그 검사는 자기
// 시계를 재는 셈이 되어 언제나 통과한다. 없거나 null이거나 문법이 다르면 거절이다.
func strictQuoteInstant(raw json.RawMessage, what string) (string, time.Time, error) {
	text, err := strictMinuteString(raw)
	if err != nil {
		return "", time.Time{}, strictQuoteRefuse(StrictQuoteReasonNoSourceInstant, what+": "+err.Error())
	}
	instant, err := strictMinuteInstant(text)
	if err != nil {
		return "", time.Time{}, strictQuoteRefuse(StrictQuoteReasonNoSourceInstant, err.Error())
	}
	return text, instant, nil
}

// strictQuoteSideTop은 한쪽 사다리의 맨 위 한 줄만 꺼낸다.
// 빈 쪽은 거절이다 — 장이 닫혀 있으면 양쪽이 빈 배열로 온다(2026-08-16 실측).
func strictQuoteSideTop(raw json.RawMessage, side string) (StrictQuoteLevel, error) {
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return StrictQuoteLevel{}, strictQuoteRefuse(StrictQuoteReasonResult, side+" is not a JSON array")
	}
	var levels []json.RawMessage
	if err := json.Unmarshal(raw, &levels); err != nil {
		return StrictQuoteLevel{}, strictQuoteRefuse(StrictQuoteReasonResult, side+": "+err.Error())
	}
	if len(levels) == 0 {
		return StrictQuoteLevel{}, strictQuoteRefuse(StrictQuoteReasonEmptySide, side+" carries no level")
	}
	fields, err := strictQuoteFields(levels[0], strictQuoteLevelKeys, StrictQuoteReasonLevel, side+"[0]")
	if err != nil {
		return StrictQuoteLevel{}, err
	}
	price, err := strictQuoteDecimal(fields["price"], side+"[0].price")
	if err != nil {
		return StrictQuoteLevel{}, err
	}
	volume, err := strictQuoteDecimal(fields["volume"], side+"[0].volume")
	if err != nil {
		return StrictQuoteLevel{}, err
	}
	return StrictQuoteLevel{Price: price, Volume: volume}, nil
}

// strictQuoteDecodeOrderbook은 호가 본문을 계약대로 읽는다. 어긋나면 거절한다.
func strictQuoteDecodeOrderbook(body []byte, currency string) (StrictQuoteLevel, StrictQuoteLevel, string, time.Time, error) {
	var ask, bid StrictQuoteLevel
	resultRaw, err := strictQuoteResult(body)
	if err != nil {
		return ask, bid, "", time.Time{}, err
	}
	result, err := strictQuoteFields(resultRaw, strictQuoteOrderbookKeys, StrictQuoteReasonResult, "result")
	if err != nil {
		return ask, bid, "", time.Time{}, err
	}
	if err := strictQuoteCheckCurrency(result["currency"], currency); err != nil {
		return ask, bid, "", time.Time{}, err
	}
	stamp, instant, err := strictQuoteInstant(result["timestamp"], "timestamp")
	if err != nil {
		return ask, bid, "", time.Time{}, err
	}
	if ask, err = strictQuoteSideTop(result["asks"], "asks"); err != nil {
		return ask, bid, "", time.Time{}, err
	}
	if bid, err = strictQuoteSideTop(result["bids"], "bids"); err != nil {
		return ask, bid, "", time.Time{}, err
	}
	return ask, bid, stamp, instant, nil
}

// strictQuoteDecodePrices는 현재가 본문을 계약대로 읽는다.
func strictQuoteDecodePrices(body []byte, symbol, currency string) (string, string, time.Time, error) {
	resultRaw, err := strictQuoteResult(body)
	if err != nil {
		return "", "", time.Time{}, err
	}
	trimmed := bytes.TrimLeft(resultRaw, " \t\r\n")
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return "", "", time.Time{}, strictQuoteRefuse(StrictQuoteReasonResult, "result is not a JSON array")
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(resultRaw, &rows); err != nil {
		return "", "", time.Time{}, strictQuoteRefuse(StrictQuoteReasonResult, "result: "+err.Error())
	}
	if len(rows) != 1 {
		return "", "", time.Time{}, strictQuoteRefuse(StrictQuoteReasonRowCount,
			"result carries "+strconv.Itoa(len(rows))+" rows for the one symbol that was asked for")
	}
	fields, err := strictQuoteFields(rows[0], strictQuotePriceRowKeys, StrictQuoteReasonRow, "row")
	if err != nil {
		return "", "", time.Time{}, err
	}
	echoed, err := strictQuoteText(fields["symbol"], StrictQuoteReasonSymbolEcho, "row.symbol")
	if err != nil {
		return "", "", time.Time{}, err
	}
	if echoed != symbol {
		return "", "", time.Time{}, strictQuoteRefuse(StrictQuoteReasonSymbolEcho,
			"row answers for "+strconv.Quote(echoed)+", not for "+strconv.Quote(symbol))
	}
	if err := strictQuoteCheckCurrency(fields["currency"], currency); err != nil {
		return "", "", time.Time{}, err
	}
	last, err := strictQuoteDecimal(fields["lastPrice"], "row.lastPrice")
	if err != nil {
		return "", "", time.Time{}, err
	}
	stamp, instant, err := strictQuoteInstant(fields["timestamp"], "row.timestamp")
	if err != nil {
		return "", "", time.Time{}, err
	}
	return last, stamp, instant, nil
}

func strictQuoteCheckCurrency(raw json.RawMessage, currency string) error {
	text, err := strictQuoteText(raw, StrictQuoteReasonCurrency, "currency")
	if err != nil {
		return err
	}
	if text != currency {
		return strictQuoteRefuse(StrictQuoteReasonCurrency,
			"currency "+strconv.Quote(text)+" is not "+currency)
	}
	return nil
}
