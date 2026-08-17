package official

// strict_minute_candles.go는 공식 1분봉 페이지를 "받은 바이트 그대로" 엄격하게 읽는다.
//
// 왜 따로 만들었나
//
// 기존 RawMinuteCandles는 표준 JSON 해독기를 쓴다. 표준 해독기는 같은 열쇠가 두 번 나오면
// 뒤엣것으로 조용히 덮어쓰고, 모르는 열쇠는 그냥 버리고, 커서가 없는 것과 null인 것과
// 빈 문자열인 것을 모두 ""로 뭉갠다. 증거로 쓰려면 그 셋은 서로 다른 사건이어야 한다.
// 그래서 이 파일은 응답 본문의 정확한 바이트를 직접 읽고, 조금이라도 계약과 다르면
// 고쳐 읽지 않고 거절한다(fail closed).
//
// 무엇을 건드리지 않나
//
// 토큰·재시도·전송은 평소 생산 경로(c.get → send → doRequest)를 그대로 쓴다. 여기서는
// ctx에 붙은 관측자를 "덧대어" 시도마다의 원본 바이트를 받아 적을 뿐이다. 이미 붙어 있던
// 관측자도 계속 호출한다(덮어쓰지 않는다).

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

// PathMinuteCandles는 1분봉 요청 경로다. 요청과 rate budget 조회가 같은 글자를 써야
// 예산이 조용히 빗나가지 않으므로 한 곳에만 적는다.
const PathMinuteCandles = "/api/v1/candles"

const (
	// 한 페이지에 담을 수 있는 봉 수의 상한(문서값).
	strictMinuteMaxCount = 200
	// 응답 본문 상한. doRequest는 상한 없이 읽으므로 여기서 읽은 뒤에 자른다.
	strictMinuteMaxBody = 2 << 20
	// 소수 문자열의 바이트 상한(문서 maxLength).
	strictMinuteMaxDecimal = 30
	// 중첩 깊이 상한. 깊이만 깊은 본문으로 스택을 무너뜨리지 못하게 한다.
	strictMinuteMaxDepth = 16
)

var (
	strictMinuteKRSymbol  = regexp.MustCompile(`^[0-9]{6}$`)
	strictMinuteUSSymbol  = regexp.MustCompile(`^[A-Z][A-Z0-9.]{0,9}$`)
	strictMinuteTimestamp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,3})?[+-]\d{2}:\d{2}$`)
)

// 봉 하나가 반드시 가져야 하는 일곱 개의 열쇠. 더도 덜도 안 된다.
var strictMinuteCandleKeys = []string{
	"timestamp", "openPrice", "highPrice", "lowPrice", "closePrice", "volume", "currency",
}

// StrictMinutePage는 한 페이지의 읽기 결과다. 생산자의 가짜 리더가 직접 만들 수 있어야
// 하므로 평범한 공개 구조체다.
type StrictMinutePage struct {
	Market string
	Symbol string
	// Candles는 받은 순서 그대로(최신이 앞) 담긴 원본 문자열이다.
	//
	// RawMinuteCandle.Timestamp는 브로커가 보낸 글자 그대로이고, 그 값은 봉이
	// **닫는** 순간이다(2026-08-18 03:29 KST 실측: 벽시계가 [03:29, 03:30) 안에 있는
	// 동안 자라던 봉의 이름표가 `03:30:00`이었다). 문서와 candle_reads.go의 주석은
	// "봉 시작 시각"이라고 말하지만 틀렸다. 이 리더는 전송 계층 DTO이므로 값을 바꾸지
	// 않는다 — 여는 시각으로 옮기는 일(−1분)은 생산자(internal/officialbars)의 몫이다.
	Candles []RawMinuteCandle
	// Terminal은 nextBefore가 null이었다는 뜻이다. 마지막 페이지다.
	Terminal bool
	// NextBefore는 해독된 커서 값이다. Terminal이면 반드시 빈 문자열이다.
	NextBefore string
	// ReadAt은 이 페이지의 본문을 다 읽은 순간이다(응답에 묶인 시각).
	ReadAt time.Time
	// StatusCode는 이 본문을 준 응답의 상태 코드다.
	StatusCode int
	// BodyDigest는 정확한 응답 바이트의 "sha256:"+16진수 지문이다.
	BodyDigest string
	// Budget은 같은 경로에서 마지막으로 관측된 rate 예산이다. 참고용이다.
	Budget RateBudget
}

// StrictMinuteCandlesError는 "계약이 어긋나서 거절했다"는 뜻이다. 전송 실패나
// classifyStatus가 만든 오류와 구분하려고 따로 둔다.
type StrictMinuteCandlesError struct {
	Reason string
	Detail string
}

func (e *StrictMinuteCandlesError) Error() string {
	return "official strict minute candles: " + e.Reason + ": " + e.Detail
}

// 거절 사유 이름표. 호출자가 어느 규칙에 걸렸는지 알 수 있게 문자열로 남긴다.
const (
	StrictReasonClientMissing = "CLIENT_MISSING"
	StrictReasonMarket        = "MARKET_INVALID"
	StrictReasonSymbol        = "SYMBOL_INVALID"
	StrictReasonCount         = "COUNT_INVALID"
	StrictReasonBefore        = "BEFORE_INVALID"
	StrictReasonNoAttempt     = "NO_SUCCESSFUL_ATTEMPT"
	StrictReasonBodyTooLarge  = "BODY_TOO_LARGE"
	StrictReasonBody          = "BODY_INVALID"
	StrictReasonResult        = "RESULT_INVALID"
	StrictReasonCandle        = "CANDLE_INVALID"
	// 봉의 시각은 반드시 분의 시작이어야 한다(결정 26).
	StrictReasonCandleNotOnMinute = "CANDLE_NOT_ON_MINUTE"
	StrictReasonInvalidUTF8       = "BODY_NOT_UTF8"
	StrictReasonOrder             = "CANDLE_ORDER_INVALID"
	StrictReasonCursor            = "CURSOR_INVALID"
)

func strictMinuteRefuse(reason, detail string) error {
	return &StrictMinuteCandlesError{Reason: reason, Detail: detail}
}

// StrictMinuteCandles는 KR/US 공통의 엄격한 1분봉 페이지 읽기다.
//
// 요청을 보내기 전에 인자 문법을 모두 확인한다. 문법이 틀린 요청은 브로커에게
// 보내지 않는다(공유 쿼터를 낭비하지 않는다).
//
// 시각 규약: 돌려주는 candle의 Timestamp도, before도, nextBefore도 모두 봉이 **닫는**
// 순간이다(2026-08-18 실측; 결정 30). before는 그 닫는 시각에 대한 **포함** 상한이므로
// `before = 지금`으로 부르면 아직 자라고 있는 봉은 저절로 빠진다(그 봉의 닫는 시각은
// 아직 오지 않았다). 이 함수는 어떤 값도 여는 시각으로 바꾸지 않는다.
func (c *Client) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (StrictMinutePage, error) {
	if c == nil {
		return StrictMinutePage{}, strictMinuteRefuse(StrictReasonClientMissing, "client is required")
	}
	currency, err := strictMinuteMarketCurrency(market)
	if err != nil {
		return StrictMinutePage{}, err
	}
	if err := strictMinuteCheckSymbol(market, symbol); err != nil {
		return StrictMinutePage{}, err
	}
	if count < 1 || count > strictMinuteMaxCount {
		return StrictMinutePage{}, strictMinuteRefuse(StrictReasonCount,
			"count must be between 1 and "+strconv.Itoa(strictMinuteMaxCount)+", got "+strconv.Itoa(count))
	}
	var beforeInstant time.Time
	if before != "" {
		beforeInstant, err = strictMinuteInstant(before)
		if err != nil {
			return StrictMinutePage{}, strictMinuteRefuse(StrictReasonBefore, err.Error())
		}
	}

	query := url.Values{}
	query.Set("symbol", symbol)
	query.Set("interval", RawCandleIntervalOneMinute)
	query.Set("count", strconv.Itoa(count))
	query.Set("adjusted", "false")
	if before != "" {
		query.Set("before", before)
	}

	// 이미 ctx에 관측자가 있으면 그것도 계속 부른다. 덮어쓰면 M0 읽기의 증거가 사라진다.
	outer, _ := ctx.Value(attemptObserverKey{}).(AttemptObserver)
	var attempts []AttemptTrace
	traced := WithAttemptObserver(ctx, func(trace AttemptTrace) {
		attempts = append(attempts, trace)
		if outer != nil {
			outer(trace)
		}
	})
	if err := c.get(traced, PathMinuteCandles, query, nil); err != nil {
		return StrictMinutePage{}, err
	}

	used, err := strictMinuteFinalAttempt(attempts)
	if err != nil {
		return StrictMinutePage{}, err
	}
	candles, terminal, cursor, err := strictMinuteDecode(used.Body, count, currency, beforeInstant)
	if err != nil {
		return StrictMinutePage{}, err
	}
	digest := sha256.Sum256(used.Body)
	return StrictMinutePage{
		Market:     market,
		Symbol:     symbol,
		Candles:    candles,
		Terminal:   terminal,
		NextBefore: cursor,
		ReadAt:     used.BodyReadComplete,
		StatusCode: used.StatusCode,
		BodyDigest: "sha256:" + hex.EncodeToString(digest[:]),
		Budget:     c.RateBudget(PathMinuteCandles),
	}, nil
}

func strictMinuteMarketCurrency(market string) (string, error) {
	switch market {
	case "KR":
		return "KRW", nil
	case "US":
		return "USD", nil
	default:
		return "", strictMinuteRefuse(StrictReasonMarket, `market must be exactly "KR" or "US", got `+strconv.Quote(market))
	}
}

func strictMinuteCheckSymbol(market, symbol string) error {
	pattern := strictMinuteUSSymbol
	if market == "KR" {
		pattern = strictMinuteKRSymbol
	}
	if !pattern.MatchString(symbol) {
		return strictMinuteRefuse(StrictReasonSymbol,
			"symbol "+strconv.Quote(symbol)+" is not canonical for market "+market)
	}
	return nil
}

// strictMinuteInstant는 문법을 먼저 보고 그 다음에 실제로 있는 시각인지 본다.
// 문법만 맞고 존재하지 않는 날짜(2월 30일 같은)는 파싱에서 걸린다.
func strictMinuteInstant(value string) (time.Time, error) {
	if !strictMinuteTimestamp.MatchString(value) {
		return time.Time{}, errors.New("timestamp " + strconv.Quote(value) +
			" must look like 2006-01-02T15:04:05[.000]+09:00 with a numeric offset")
	}
	instant, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.New("timestamp " + strconv.Quote(value) + " does not exist")
	}
	return instant, nil
}

// strictMinuteFinalAttempt는 마지막 시도를 고르고, 그것이 2xx가 아니면 거절한다.
// "2xx인 시도들 중 마지막"이 아니라 "마지막 시도"인 것이 중요하다. 앞의 성공을 뒤의
// 실패가 덮어썼는데도 앞의 본문을 증거로 쓰는 일이 생기면 안 되기 때문이다.
// 오늘의 send는 2xx 뒤에 다시 시도하지 않으므로 두 규칙은 같은 결과를 낸다(결정 28).
func strictMinuteFinalAttempt(attempts []AttemptTrace) (AttemptTrace, error) {
	if len(attempts) == 0 {
		// 도달 불가 방어: get이 nil을 돌려줬다면 doRequest가 적어도 한 번은 관측을 냈다.
		return AttemptTrace{}, strictMinuteRefuse(StrictReasonNoAttempt, "no attempt was observed")
	}
	last := attempts[len(attempts)-1]
	if last.Err != nil || last.StatusCode < 200 || last.StatusCode >= 300 {
		// 도달 불가 방어: 마지막 시도가 2xx가 아니면 get이 이미 오류를 돌려준다.
		return AttemptTrace{}, strictMinuteRefuse(StrictReasonNoAttempt,
			"the last attempt returned status "+strconv.Itoa(last.StatusCode))
	}
	return last, nil
}

// strictMinuteDecode는 본문 바이트를 계약대로 읽는다. 어긋나면 고치지 않고 거절한다.
func strictMinuteDecode(body []byte, count int, currency string, beforeInstant time.Time) ([]RawMinuteCandle, bool, string, error) {
	if len(body) > strictMinuteMaxBody {
		return nil, false, "", strictMinuteRefuse(StrictReasonBodyTooLarge,
			"body is "+strconv.Itoa(len(body))+" bytes, above the cap of "+strconv.Itoa(strictMinuteMaxBody))
	}
	// 표준 해독기는 깨진 바이트를 U+FFFD로 조용히 바꾼다. 그러면 브로커가 보내지 않은
	// 값이 증거가 되므로, 바이트를 보기 전에 먼저 거절한다.
	if !utf8.Valid(body) {
		return nil, false, "", strictMinuteRefuse(StrictReasonInvalidUTF8, "body is not valid UTF-8")
	}
	if err := strictMinuteCheckJSON(body); err != nil {
		return nil, false, "", strictMinuteRefuse(StrictReasonBody, err.Error())
	}
	envelope, err := strictMinuteObject(body)
	if err != nil {
		return nil, false, "", strictMinuteRefuse(StrictReasonBody, "envelope: "+err.Error())
	}
	// 봉투의 다른 열쇠는 무시한다. result만이 계약이다.
	resultRaw, found := envelope["result"]
	if !found {
		return nil, false, "", strictMinuteRefuse(StrictReasonResult, "envelope has no 'result' key")
	}
	result, err := strictMinuteObject(resultRaw)
	if err != nil {
		return nil, false, "", strictMinuteRefuse(StrictReasonResult, err.Error())
	}
	candlesRaw, hasCandles := result["candles"]
	cursorRaw, hasCursor := result["nextBefore"]
	if !hasCandles {
		return nil, false, "", strictMinuteRefuse(StrictReasonResult, "result has no 'candles' key")
	}
	if len(result) != 2 {
		return nil, false, "", strictMinuteRefuse(StrictReasonResult,
			"result must carry exactly 'candles' and 'nextBefore'")
	}

	candles, instants, err := strictMinuteCandles(candlesRaw, count, currency, beforeInstant)
	if err != nil {
		return nil, false, "", err
	}
	terminal, cursor, err := strictMinuteCursor(cursorRaw, hasCursor, instants)
	if err != nil {
		return nil, false, "", err
	}
	return candles, terminal, cursor, nil
}

func strictMinuteCandles(raw []byte, count int, currency string, beforeInstant time.Time) ([]RawMinuteCandle, []time.Time, error) {
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, nil, strictMinuteRefuse(StrictReasonResult, "candles is not a JSON array")
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, nil, strictMinuteRefuse(StrictReasonResult, "candles: "+err.Error())
	}
	if len(rows) > count {
		return nil, nil, strictMinuteRefuse(StrictReasonCandle,
			"page carries "+strconv.Itoa(len(rows))+" candles for a count of "+strconv.Itoa(count))
	}
	candles := make([]RawMinuteCandle, 0, len(rows))
	instants := make([]time.Time, 0, len(rows))
	for index, row := range rows {
		candle, instant, err := strictMinuteCandle(row, currency)
		if err != nil {
			return nil, nil, err
		}
		if index > 0 && !instant.Before(instants[index-1]) {
			return nil, nil, strictMinuteRefuse(StrictReasonOrder,
				"candle "+strconv.Itoa(index)+" is not strictly older than the one before it")
		}
		if !beforeInstant.IsZero() && instant.After(beforeInstant) {
			return nil, nil, strictMinuteRefuse(StrictReasonOrder,
				"candle "+strconv.Itoa(index)+" is newer than the inclusive upper bound")
		}
		candles = append(candles, candle)
		instants = append(instants, instant)
	}
	return candles, instants, nil
}

func strictMinuteCandle(raw []byte, currency string) (RawMinuteCandle, time.Time, error) {
	fields, err := strictMinuteObject(raw)
	if err != nil {
		return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle, err.Error())
	}
	// 열쇠는 정확히 일곱 개여야 한다. 하나라도 더 있으면 브로커가 모양을 바꿨다는 뜻이고,
	// 모르는 값을 조용히 버리는 대신 눈에 보이게 실패하는 쪽이 안전하다.
	if len(fields) != len(strictMinuteCandleKeys) {
		return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle,
			"candle carries "+strconv.Itoa(len(fields))+" keys, not the "+
				strconv.Itoa(len(strictMinuteCandleKeys))+" the contract fixes")
	}
	values := make(map[string]string, len(strictMinuteCandleKeys))
	for _, key := range strictMinuteCandleKeys {
		value, found := fields[key]
		if !found {
			return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle,
				"candle has no "+strconv.Quote(key)+" key")
		}
		text, err := strictMinuteString(value)
		if err != nil {
			return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle, key+": "+err.Error())
		}
		values[key] = text
	}
	for _, key := range []string{"openPrice", "highPrice", "lowPrice", "closePrice", "volume"} {
		// 숫자 문법은 여기서 보지 않는다. 소수의 유일한 권위는 strategyevidence다.
		// 여기서는 "문자열이고, 비어 있지 않고, 문서가 정한 길이 안"인지만 본다.
		if len(values[key]) > strictMinuteMaxDecimal {
			return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle,
				key+" is longer than "+strconv.Itoa(strictMinuteMaxDecimal)+" bytes")
		}
	}
	if values["currency"] != currency {
		return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle,
			"currency "+strconv.Quote(values["currency"])+" is not "+currency)
	}
	instant, err := strictMinuteInstant(values["timestamp"])
	if err != nil {
		return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandle, err.Error())
	}
	// 1분봉의 시각표는 언제나 분 경계에 놓인다(실측 계약). 초나 소수가 붙어 있으면
	// 거절한다. 여기서 막지 않으면 어긋난 봉 하나가 바로 아래 봉의 successor 주장까지
	// 망가뜨린다. (이 값은 봉이 닫는 순간이다 — 위 함수 주석 참고.)
	if instant.Second() != 0 || instant.Nanosecond() != 0 {
		return RawMinuteCandle{}, time.Time{}, strictMinuteRefuse(StrictReasonCandleNotOnMinute,
			"timestamp "+strconv.Quote(values["timestamp"])+" does not start a minute")
	}
	return RawMinuteCandle{
		Timestamp: values["timestamp"],
		Open:      values["openPrice"],
		High:      values["highPrice"],
		Low:       values["lowPrice"],
		Close:     values["closePrice"],
		Volume:    values["volume"],
		Currency:  values["currency"],
	}, instant, nil
}

// strictMinuteCursor는 커서의 세 가지 상태를 구분한다.
// 문자열이면 다음 페이지가 있고, null이면 마지막이며, 그 밖의 어떤 모양도 거절이다.
func strictMinuteCursor(raw []byte, present bool, instants []time.Time) (bool, string, error) {
	if !present {
		return false, "", strictMinuteRefuse(StrictReasonCursor,
			"result has no 'nextBefore' key; absent is not terminal")
	}
	trimmed := bytes.TrimSpace(raw)
	if bytes.Equal(trimmed, []byte("null")) {
		return true, "", nil
	}
	if len(trimmed) == 0 || trimmed[0] != '"' {
		return false, "", strictMinuteRefuse(StrictReasonCursor, "nextBefore must be a JSON string or null")
	}
	value, err := strictMinuteString(trimmed)
	if err != nil {
		return false, "", strictMinuteRefuse(StrictReasonCursor, err.Error())
	}
	cursor, err := strictMinuteInstant(value)
	if err != nil {
		return false, "", strictMinuteRefuse(StrictReasonCursor, err.Error())
	}
	if len(instants) > 0 && !cursor.Before(instants[len(instants)-1]) {
		return false, "", strictMinuteRefuse(StrictReasonCursor,
			"nextBefore is not strictly older than the oldest bar of this page")
	}
	return false, value, nil
}

// strictMinuteString은 JSON 문자열 하나를 꺼낸다. 숫자·null·객체는 전부 거절이다.
func strictMinuteString(raw []byte) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '"' {
		return "", errors.New("value is not a JSON string")
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return "", errors.New("value is not a JSON string")
	}
	if value == "" {
		return "", errors.New("value is an empty string")
	}
	return value, nil
}

// strictMinuteObject는 객체 하나의 열쇠와 원본 값을 그대로 담아 온다.
//
// 중복 열쇠와 후행 값은 여기서 다시 보지 않는다. strictMinuteCheckJSON이 본문 전체를
// 한 번 훑으며 어느 깊이에서든 이미 거절했기 때문이다. 같은 규칙을 두 곳에 두면
// 한쪽을 지워도 다른 쪽이 가려 주어, 어느 쪽도 시험으로 확인할 수 없게 된다.
func strictMinuteObject(source []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	opening, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := opening.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("value is not an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("object key is not a string")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[key] = bytes.Clone(value)
	}
	closing, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := closing.(json.Delim); !ok || delimiter != '}' {
		return nil, errors.New("object did not close")
	}
	return fields, nil
}

// strictMinuteCheckJSON은 본문 전체를 한 번 훑으며 어느 깊이에서든 중복 열쇠를 거절한다.
// strictMinuteObject는 자기 층만 보므로, 봉 안쪽까지 보장하려면 이 훑기가 필요하다.
func strictMinuteCheckJSON(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := strictMinuteWalk(decoder, 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func strictMinuteWalk(decoder *json.Decoder, depth int) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	// 깊이는 상자(객체·배열)만 센다. 봉투 자신이 0층이므로 그 안에 17겹이 들어오면 거절이다.
	if depth > strictMinuteMaxDepth {
		return errors.New("JSON nesting is deeper than " + strconv.Itoa(strictMinuteMaxDepth))
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if err := strictMinuteWalk(decoder, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := strictMinuteWalk(decoder, depth+1); err != nil {
				return err
			}
		}
	default:
		// 도달 불가 방어: encoding/json이 내는 구분자는 '{' '[' ']' '}' 넷뿐이고
		// 닫는 것은 More()/Token() 짝에서 소비된다.
		return errors.New("unexpected JSON delimiter")
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	return nil
}
