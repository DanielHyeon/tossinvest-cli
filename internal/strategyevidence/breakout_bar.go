package strategyevidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// breakout 전략이 쓰는 공식 증거 두 가지다.
// 하나는 "이미 닫힌 1분봉", 다른 하나는 "호가창 맨 위 한 줄(L1)"이다.
const (
	KindOfficialClosedBar1m EvidenceKind = "official_closed_bar_1m"
	KindOfficialQuoteL1     EvidenceKind = "official_quote_l1"
)

// 스키마 이름은 페이로드 안에도 들어간다. 나중에 모양이 바뀌면 이름이 먼저 바뀌므로
// 옛 바이트를 새 규칙으로 잘못 읽는 일이 생기지 않는다.
const (
	SchemaOfficialClosedBar1m = "official_closed_bar_1m:v1"
	SchemaOfficialQuoteL1     = "official_quote_l1:v1"
)

const (
	// 1분봉이므로 간격은 언제나 60000밀리초다.
	ClosedBar1mIntervalMS = 60000
	// 봉의 시각표는 "봉이 시작한 시각"이다 (a112 결정 7).
	BarLabelOpenAt = "open_at"
	// 다음 봉을 실제로 본 뒤에만 이 봉을 닫힌 것으로 인정한다 (a112 결정 6).
	BarFinalitySuccessorObserved = "successor_observed"
	// 봉 발행자 이름표의 판 번호.
	BarIssuerMappingVersion = "a112-bar-issuer-v1"
)

const maxUint64 = ^uint64(0)

// RawClosedBar1m은 브로커가 보낸 소수 문자열을 그대로 담는다. 실수(float)로 바꾸지 않는다.
type RawClosedBar1m struct {
	Open, High, Low, Close, Volume string
}

// ClosedBar1mPayload는 닫힌 1분봉 한 개의 증거 본문이다.
// 값은 모두 정수 minor 단위이고, 원본 문자열은 Raw에 그대로 남는다.
type ClosedBar1mPayload struct {
	Schema               string
	Market               string
	Symbol               string
	SessionID            string
	CalendarVersion      string
	IntervalMS           uint64
	BarLabel             string
	OpenAtMS             uint64
	Finality             string
	Closed               bool
	RegularSession       bool
	Currency             string
	PriceScale           uint64
	OpenMinor            uint64
	HighMinor            uint64
	LowMinor             uint64
	CloseMinor           uint64
	Volume               uint64
	Raw                  RawClosedBar1m
	Revision             uint64
	SuccessorOpenAtMS    uint64
	SourceObservedAtMS   uint64
	SourceResponseDigest string
}

// RawQuoteL1은 호가 한 줄의 원본 소수 문자열이다.
type RawQuoteL1 struct {
	Bid, Ask, Last string
}

// QuoteL1Payload는 호가창 맨 위 한 줄의 증거 본문이다.
type QuoteL1Payload struct {
	Schema               string
	Market               string
	Symbol               string
	Currency             string
	PriceScale           uint64
	BidMinor             uint64
	AskMinor             uint64
	LastMinor            uint64
	Raw                  RawQuoteL1
	SourceObservedAtMS   uint64
	ReceivedAtMS         uint64
	SourceResponseDigest string
	Revision             uint64
}

// DecodeClosedBar1mPayload는 저장된 정규 바이트를 다시 엄격하게 읽는다.
// 저장할 때와 읽을 때가 같은 규칙을 쓰므로, 한 번 통과한 바이트만 다시 통과한다.
func DecodeClosedBar1mPayload(canonical []byte) (ClosedBar1mPayload, error) {
	object, err := decodeCanonicalObject(canonical)
	if err != nil {
		return ClosedBar1mPayload{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	if err := rejectSecretFields(object); err != nil {
		return ClosedBar1mPayload{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	bar, err := decodeClosedBar1mObject(object)
	if err != nil {
		return ClosedBar1mPayload{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	return bar, nil
}

// DecodeQuoteL1Payload는 저장된 호가 증거를 다시 엄격하게 읽는다.
func DecodeQuoteL1Payload(canonical []byte) (QuoteL1Payload, error) {
	object, err := decodeCanonicalObject(canonical)
	if err != nil {
		return QuoteL1Payload{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	if err := rejectSecretFields(object); err != nil {
		return QuoteL1Payload{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	quote, err := decodeQuoteL1Object(object)
	if err != nil {
		return QuoteL1Payload{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	return quote, nil
}

// newClosedBar1mHeader는 봉 하나의 출처 머리말을 만든다.
// 같은 봉과 같은 판(revision)이면 언제나 똑같은 머리말이 나온다.
func newClosedBar1mHeader(market marketclock.Market, symbol, sessionID, calendarVersion string,
	openAt time.Time, revision uint64, observedAt time.Time, currency string) (Header, error) {
	normalizedMarket, code, err := marketAndCode(market)
	if err != nil {
		return Header{}, err
	}
	upperSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if !canonicalSymbolText(upperSymbol) {
		return Header{}, invalid(RefusalIdentityMismatch, "symbol", "a canonical symbol without the ':' record separator is required")
	}
	trimmedSession := strings.TrimSpace(sessionID)
	sessionDate, err := sessionDateFor(code, trimmedSession)
	if err != nil {
		return Header{}, invalid(RefusalIdentityMismatch, "session_id", err.Error())
	}
	if !canonicalPayloadText(calendarVersion) {
		return Header{}, invalid(RefusalIdentityMismatch, "calendar_version", "a canonical calendar version is required")
	}
	if err := checkMarketCurrency(code, currency); err != nil {
		return Header{}, err
	}
	if revision == 0 {
		return Header{}, invalid(RefusalIdentityMismatch, "revision", "revisions start at 1")
	}
	if openAt.IsZero() || observedAt.IsZero() {
		return Header{}, invalid(RefusalTimestampInvalid, "timestamps", "bar open and observation instants are required")
	}
	openAtUTC := openAt.UTC()
	openAtMS := openAtUTC.UnixMilli()
	if openAtMS <= 0 || openAtMS%ClosedBar1mIntervalMS != 0 {
		return Header{}, invalid(RefusalTimestampInvalid, "open_at", "a 1 minute bar must open on a whole minute")
	}
	tradingDay, err := normalizedMarket.TradingDay(openAtUTC)
	if err != nil {
		return Header{}, invalid(RefusalTimestampInvalid, "open_at", err.Error())
	}
	if tradingDay != sessionDate {
		return Header{}, invalid(RefusalIdentityMismatch, "session_id",
			"the bar opens on market-local trading day "+tradingDay+", not on the session date "+sessionDate)
	}
	closedAt := openAtUTC.Add(time.Minute)
	observedUTC := observedAt.UTC()
	if observedUTC.Before(closedAt) {
		return Header{}, invalid(RefusalTimestampInvalid, "observed_at", "a bar becomes evidence only after it closed")
	}
	recordID := closedBar1mRecordID(code, upperSymbol, trimmedSession, ClosedBar1mIntervalMS, uint64(openAtMS))
	return Header{
		EvidenceID:                 evidenceIDFor(recordID, revision),
		Market:                     normalizedMarket,
		Symbol:                     upperSymbol,
		IssuerIdentity:             code + ":" + upperSymbol,
		IssuerMappingVersion:       BarIssuerMappingVersion,
		Kind:                       KindOfficialClosedBar1m,
		SchemaVersion:              SchemaOfficialClosedBar1m,
		Authority:                  AuthorityTossOpenAPI,
		SourceRecordID:             recordID,
		RevisionIdentity:           revisionIdentityFor(revision),
		SupersedesRevisionIdentity: supersededRevisionIdentity(revision),
		MarketEffectiveDate:        sessionDate,
		SourceEventAt:              openAtUTC,
		SourceAvailableAt:          closedAt,
		ObservedAt:                 observedUTC,
		IngestedAt:                 observedUTC,
		Currency:                   strings.ToUpper(strings.TrimSpace(currency)),
		Unit:                       "minor",
		Availability:               AvailabilityAvailable,
		Confidence:                 ConfidenceVerified,
	}, nil
}

// newQuoteL1Header는 호가 한 줄의 출처 머리말을 만든다.
// 호가는 세션이 아니라 "본 순간"이 신원이므로 기록 번호에 관측 시각이 들어간다.
func newQuoteL1Header(market marketclock.Market, symbol string, sourceObservedAt time.Time,
	revision uint64, receivedAt time.Time, currency string) (Header, error) {
	normalizedMarket, code, err := marketAndCode(market)
	if err != nil {
		return Header{}, err
	}
	upperSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if !canonicalSymbolText(upperSymbol) {
		return Header{}, invalid(RefusalIdentityMismatch, "symbol", "a canonical symbol without the ':' record separator is required")
	}
	if err := checkMarketCurrency(code, currency); err != nil {
		return Header{}, err
	}
	if revision == 0 {
		return Header{}, invalid(RefusalIdentityMismatch, "revision", "revisions start at 1")
	}
	if sourceObservedAt.IsZero() || receivedAt.IsZero() {
		return Header{}, invalid(RefusalTimestampInvalid, "timestamps", "observation and receipt instants are required")
	}
	observedUTC := sourceObservedAt.UTC()
	receivedUTC := receivedAt.UTC()
	if receivedUTC.Before(observedUTC) {
		return Header{}, invalid(RefusalTimestampInvalid, "received_at", "a quote cannot be received before it was observed")
	}
	observedMS := observedUTC.UnixMilli()
	if observedMS <= 0 {
		return Header{}, invalid(RefusalTimestampInvalid, "source_observed_at", "the observation instant must be a positive unix millisecond")
	}
	effectiveDate, err := normalizedMarket.TradingDay(observedUTC)
	if err != nil {
		return Header{}, invalid(RefusalTimestampInvalid, "market_effective_date", err.Error())
	}
	recordID := strings.Join([]string{code, upperSymbol, "quote_l1", formatUint(uint64(observedMS))}, ":")
	return Header{
		EvidenceID:                 evidenceIDFor(recordID, revision),
		Market:                     normalizedMarket,
		Symbol:                     upperSymbol,
		IssuerIdentity:             code + ":" + upperSymbol,
		IssuerMappingVersion:       BarIssuerMappingVersion,
		Kind:                       KindOfficialQuoteL1,
		SchemaVersion:              SchemaOfficialQuoteL1,
		Authority:                  AuthorityTossOpenAPI,
		SourceRecordID:             recordID,
		RevisionIdentity:           revisionIdentityFor(revision),
		SupersedesRevisionIdentity: supersededRevisionIdentity(revision),
		MarketEffectiveDate:        effectiveDate,
		SourceEventAt:              observedUTC,
		SourceAvailableAt:          observedUTC,
		ObservedAt:                 receivedUTC,
		IngestedAt:                 receivedUTC,
		Currency:                   strings.ToUpper(strings.TrimSpace(currency)),
		Unit:                       "minor",
		Availability:               AvailabilityAvailable,
		Confidence:                 ConfidenceVerified,
	}, nil
}

// ---- 생산자용 결합 생성자 ----

// ClosedBar1mInput은 봉 하나를 증거로 만들 때 생산자가 주는 값 전부다.
// 가격은 브로커가 준 소수 문자열만 준다. 정수 minor 값과 price_scale은 여기서 만든다.
type ClosedBar1mInput struct {
	Market               marketclock.Market
	Symbol               string
	SessionID            string
	CalendarVersion      string
	OpenAt               time.Time
	Revision             uint64
	ObservedAt           time.Time
	Currency             string
	Raw                  RawClosedBar1m
	SuccessorOpenAt      time.Time
	RegularSession       bool
	SourceResponseDigest string
}

// QuoteL1Input은 호가 한 줄을 증거로 만들 때 생산자가 주는 값 전부다.
type QuoteL1Input struct {
	Market               marketclock.Market
	Symbol               string
	SourceObservedAt     time.Time
	ReceivedAt           time.Time
	Revision             uint64
	Currency             string
	Raw                  RawQuoteL1
	SourceResponseDigest string
}

// NewClosedBar1mEnvelope는 머리말과 본문을 한 번에 만든다.
// 시장·종목·통화·판 번호·시각이 한 곳에서만 읽히므로, 생산자가 둘을 서로 다르게 적을 수 없다.
// SuccessorOpenAt에는 "이 봉이 닫혔음을 증명한, 실제로 관측한 다음 봉"의 여는 시각을 넣는다.
// L1a는 그 값이 구조적으로 말이 되는지만 본다(분 단위·이 봉보다 뒤·관측 시각 이전).
// 정말로 그 봉을 봤는지는 부르는 쪽(L1b 생산자)의 의무다 (a112 결정 6, successor_observed).
func NewClosedBar1mEnvelope(input ClosedBar1mInput) (Envelope, error) {
	header, err := newClosedBar1mHeader(input.Market, input.Symbol, input.SessionID, input.CalendarVersion,
		input.OpenAt, input.Revision, input.ObservedAt, input.Currency)
	if err != nil {
		return Envelope{}, err
	}
	_, code, err := marketAndCode(input.Market)
	if err != nil {
		return Envelope{}, err
	}
	if input.SuccessorOpenAt.IsZero() {
		return Envelope{}, invalid(RefusalTimestampInvalid, "successor_open_at",
			"the observed successor bar instant is required (a112 decision 6)")
	}
	successorMS := input.SuccessorOpenAt.UTC().UnixMilli()
	if successorMS <= 0 {
		return Envelope{}, invalid(RefusalTimestampInvalid, "successor_open_at", "must be a positive unix millisecond")
	}
	currency, scale, _ := marketMoney(code)
	minors, err := recomputeMinors(KindOfficialClosedBar1m, []rawMinorInput{
		{"raw.open", input.Raw.Open, scale},
		{"raw.high", input.Raw.High, scale},
		{"raw.low", input.Raw.Low, scale},
		{"raw.close", input.Raw.Close, scale},
		{"raw.volume", input.Raw.Volume, 0},
	})
	if err != nil {
		return Envelope{}, err
	}
	payload, err := json.Marshal(map[string]any{
		"schema":           SchemaOfficialClosedBar1m,
		"market":           code,
		"symbol":           header.Symbol,
		"session_id":       strings.TrimSpace(input.SessionID),
		"calendar_version": input.CalendarVersion,
		"interval_ms":      uint64(ClosedBar1mIntervalMS),
		"bar_label":        BarLabelOpenAt,
		"open_at_ms":       uint64(header.SourceEventAt.UnixMilli()),
		"finality":         BarFinalitySuccessorObserved,
		"closed":           true,
		"regular_session":  input.RegularSession,
		"currency":         currency,
		"price_scale":      scale,
		"open_minor":       minors[0],
		"high_minor":       minors[1],
		"low_minor":        minors[2],
		"close_minor":      minors[3],
		"volume":           minors[4],
		"raw": map[string]any{
			"open": input.Raw.Open, "high": input.Raw.High, "low": input.Raw.Low,
			"close": input.Raw.Close, "volume": input.Raw.Volume,
		},
		"revision":               input.Revision,
		"successor_open_at_ms":   uint64(successorMS),
		"source_observed_at_ms":  uint64(header.ObservedAt.UnixMilli()),
		"source_response_digest": input.SourceResponseDigest,
	})
	if err != nil {
		return Envelope{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	return NewEnvelope(header, payload)
}

// NewQuoteL1Envelope는 호가 증거의 머리말과 본문을 한 번에 만든다.
func NewQuoteL1Envelope(input QuoteL1Input) (Envelope, error) {
	header, err := newQuoteL1Header(input.Market, input.Symbol, input.SourceObservedAt,
		input.Revision, input.ReceivedAt, input.Currency)
	if err != nil {
		return Envelope{}, err
	}
	_, code, err := marketAndCode(input.Market)
	if err != nil {
		return Envelope{}, err
	}
	currency, scale, _ := marketMoney(code)
	minors, err := recomputeMinors(KindOfficialQuoteL1, []rawMinorInput{
		{"raw.bid", input.Raw.Bid, scale},
		{"raw.ask", input.Raw.Ask, scale},
		{"raw.last", input.Raw.Last, scale},
	})
	if err != nil {
		return Envelope{}, err
	}
	payload, err := json.Marshal(map[string]any{
		"schema":                 SchemaOfficialQuoteL1,
		"market":                 code,
		"symbol":                 header.Symbol,
		"currency":               currency,
		"price_scale":            scale,
		"bid_minor":              minors[0],
		"ask_minor":              minors[1],
		"last_minor":             minors[2],
		"raw":                    map[string]any{"bid": input.Raw.Bid, "ask": input.Raw.Ask, "last": input.Raw.Last},
		"source_observed_at_ms":  uint64(header.SourceEventAt.UnixMilli()),
		"received_at_ms":         uint64(header.ObservedAt.UnixMilli()),
		"source_response_digest": input.SourceResponseDigest,
		"revision":               input.Revision,
	})
	if err != nil {
		return Envelope{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	return NewEnvelope(header, payload)
}

type rawMinorInput struct {
	field string
	raw   string
	scale uint64
}

func recomputeMinors(kind EvidenceKind, fields []rawMinorInput) ([]uint64, error) {
	values := make([]uint64, 0, len(fields))
	for _, field := range fields {
		value, err := minorFromRawDecimal(field.raw, field.scale)
		if err != nil {
			return nil, invalid(RefusalPayloadInvalid, "payload", payloadFieldError(kind, field.field, err.Error()).Error())
		}
		values = append(values, value)
	}
	return values, nil
}

// ---- 엄격한 해독기 ----

func decodeClosedBar1mObject(object map[string]any) (ClosedBar1mPayload, error) {
	reader := newPayloadReader(KindOfficialClosedBar1m, object)
	var bar ClosedBar1mPayload
	bar.Schema = reader.text("schema")
	bar.Market = reader.text("market")
	bar.Symbol = reader.text("symbol")
	bar.SessionID = reader.text("session_id")
	bar.CalendarVersion = reader.text("calendar_version")
	bar.IntervalMS = reader.integer("interval_ms")
	bar.BarLabel = reader.text("bar_label")
	bar.OpenAtMS = reader.integer("open_at_ms")
	bar.Finality = reader.text("finality")
	bar.Closed = reader.flag("closed")
	bar.RegularSession = reader.flag("regular_session")
	bar.Currency = reader.text("currency")
	bar.PriceScale = reader.integer("price_scale")
	bar.OpenMinor = reader.integer("open_minor")
	bar.HighMinor = reader.integer("high_minor")
	bar.LowMinor = reader.integer("low_minor")
	bar.CloseMinor = reader.integer("close_minor")
	bar.Volume = reader.integer("volume")
	raw := reader.child("raw")
	bar.Raw.Open = raw.text("open")
	bar.Raw.High = raw.text("high")
	bar.Raw.Low = raw.text("low")
	bar.Raw.Close = raw.text("close")
	bar.Raw.Volume = raw.text("volume")
	_ = raw.done()
	bar.Revision = reader.integer("revision")
	bar.SuccessorOpenAtMS = reader.integer("successor_open_at_ms")
	bar.SourceObservedAtMS = reader.integer("source_observed_at_ms")
	bar.SourceResponseDigest = reader.text("source_response_digest")
	if err := reader.done(); err != nil {
		return ClosedBar1mPayload{}, err
	}
	if err := checkClosedBar1m(bar); err != nil {
		return ClosedBar1mPayload{}, err
	}
	return bar, nil
}

func checkClosedBar1m(bar ClosedBar1mPayload) error {
	kind := KindOfficialClosedBar1m
	if bar.Schema != SchemaOfficialClosedBar1m {
		return payloadFieldError(kind, "schema", "must be "+strconv.Quote(SchemaOfficialClosedBar1m))
	}
	currency, scale, known := marketMoney(bar.Market)
	if !known {
		return payloadFieldError(kind, "market", `must be "KR" or "US"`)
	}
	if bar.Currency != currency {
		return payloadFieldError(kind, "currency", "must be "+currency+" in market "+bar.Market)
	}
	if bar.PriceScale != scale {
		return payloadFieldError(kind, "price_scale", "must be "+formatUint(scale)+" for "+currency)
	}
	if !canonicalStoredSymbolText(bar.Symbol) {
		return payloadFieldError(kind, "symbol", "must be a canonical upper-case symbol without the ':' record separator")
	}
	sessionDate, err := sessionDateFor(bar.Market, bar.SessionID)
	if err != nil {
		return payloadFieldError(kind, "session_id", err.Error())
	}
	if !canonicalPayloadText(bar.CalendarVersion) {
		return payloadFieldError(kind, "calendar_version", "must be a canonical non-empty string")
	}
	if bar.IntervalMS != ClosedBar1mIntervalMS {
		return payloadFieldError(kind, "interval_ms", "must be "+formatUint(ClosedBar1mIntervalMS))
	}
	if bar.BarLabel != BarLabelOpenAt {
		return payloadFieldError(kind, "bar_label", "must be "+strconv.Quote(BarLabelOpenAt))
	}
	if bar.OpenAtMS == 0 || bar.OpenAtMS%ClosedBar1mIntervalMS != 0 {
		return payloadFieldError(kind, "open_at_ms", "must be a positive unix millisecond on a whole minute")
	}
	if bar.Finality != BarFinalitySuccessorObserved {
		return payloadFieldError(kind, "finality", "must be "+strconv.Quote(BarFinalitySuccessorObserved))
	}
	if !bar.Closed {
		return payloadFieldError(kind, "closed", "must be true; an unfinished bar is not evidence")
	}
	// 관측 시각은 적어도 봉이 끝난 시각 이후여야 한다. 미래 봉 거절은 이 규칙의 부분집합이다.
	if bar.SourceObservedAtMS < bar.OpenAtMS+bar.IntervalMS {
		return payloadFieldError(kind, "source_observed_at_ms",
			"must be at or after open_at_ms + interval_ms; a bar that has not finished is not evidence")
	}
	// 봉은 자기 세션의 달력 하루 안에 있어야 한다. 날짜는 시장 현지 시각으로 센다.
	market, err := marketclock.ParseMarket(bar.Market)
	if err != nil {
		return payloadFieldError(kind, "market", `must be "KR" or "US"`)
	}
	barDay, err := market.TradingDay(time.UnixMilli(int64(bar.OpenAtMS)).UTC())
	if err != nil {
		return payloadFieldError(kind, "open_at_ms", err.Error())
	}
	if barDay != sessionDate {
		return payloadFieldError(kind, "session_id",
			"must carry the market-local trading day of open_at_ms; the bar opens on "+barDay)
	}
	for _, minor := range []struct {
		field string
		value uint64
	}{{"open_minor", bar.OpenMinor}, {"high_minor", bar.HighMinor}, {"low_minor", bar.LowMinor}, {"close_minor", bar.CloseMinor}} {
		if minor.value == 0 {
			return payloadFieldError(kind, minor.field, "must be at least 1")
		}
	}
	if bar.LowMinor > bar.HighMinor {
		return payloadFieldError(kind, "low_minor", "must not be above high_minor")
	}
	if bar.OpenMinor < bar.LowMinor || bar.OpenMinor > bar.HighMinor {
		return payloadFieldError(kind, "open_minor", "must sit between low_minor and high_minor")
	}
	if bar.CloseMinor < bar.LowMinor || bar.CloseMinor > bar.HighMinor {
		return payloadFieldError(kind, "close_minor", "must sit between low_minor and high_minor")
	}
	for _, pair := range []struct {
		field, rawField, raw string
		declared             uint64
		scale                uint64
	}{
		{"open_minor", "raw.open", bar.Raw.Open, bar.OpenMinor, scale},
		{"high_minor", "raw.high", bar.Raw.High, bar.HighMinor, scale},
		{"low_minor", "raw.low", bar.Raw.Low, bar.LowMinor, scale},
		{"close_minor", "raw.close", bar.Raw.Close, bar.CloseMinor, scale},
		{"volume", "raw.volume", bar.Raw.Volume, bar.Volume, 0},
	} {
		if err := checkRawMinor(kind, pair.field, pair.rawField, pair.raw, pair.scale, pair.declared); err != nil {
			return err
		}
	}
	if bar.Revision == 0 {
		return payloadFieldError(kind, "revision", "must be at least 1")
	}
	// 뒤따르는 봉을 실제로 봤다는 주장. 여기서는 구조적으로 말이 되는지만 본다.
	if bar.SuccessorOpenAtMS%ClosedBar1mIntervalMS != 0 {
		return payloadFieldError(kind, "successor_open_at_ms", "must be a unix millisecond on a whole minute")
	}
	if bar.SuccessorOpenAtMS < bar.OpenAtMS+bar.IntervalMS {
		return payloadFieldError(kind, "successor_open_at_ms", "must be at least one interval after open_at_ms")
	}
	if bar.SuccessorOpenAtMS > bar.SourceObservedAtMS {
		return payloadFieldError(kind, "successor_open_at_ms", "cannot be later than the instant the successor was observed")
	}
	if !canonicalSHA256Digest(bar.SourceResponseDigest) {
		return payloadFieldError(kind, "source_response_digest", "must be sha256: followed by 64 lowercase hex digits")
	}
	return nil
}

func decodeQuoteL1Object(object map[string]any) (QuoteL1Payload, error) {
	reader := newPayloadReader(KindOfficialQuoteL1, object)
	var quote QuoteL1Payload
	quote.Schema = reader.text("schema")
	quote.Market = reader.text("market")
	quote.Symbol = reader.text("symbol")
	quote.Currency = reader.text("currency")
	quote.PriceScale = reader.integer("price_scale")
	quote.BidMinor = reader.integer("bid_minor")
	quote.AskMinor = reader.integer("ask_minor")
	quote.LastMinor = reader.integer("last_minor")
	raw := reader.child("raw")
	quote.Raw.Bid = raw.text("bid")
	quote.Raw.Ask = raw.text("ask")
	quote.Raw.Last = raw.text("last")
	_ = raw.done()
	quote.SourceObservedAtMS = reader.integer("source_observed_at_ms")
	quote.ReceivedAtMS = reader.integer("received_at_ms")
	quote.SourceResponseDigest = reader.text("source_response_digest")
	quote.Revision = reader.integer("revision")
	if err := reader.done(); err != nil {
		return QuoteL1Payload{}, err
	}
	if err := checkQuoteL1(quote); err != nil {
		return QuoteL1Payload{}, err
	}
	return quote, nil
}

func checkQuoteL1(quote QuoteL1Payload) error {
	kind := KindOfficialQuoteL1
	if quote.Schema != SchemaOfficialQuoteL1 {
		return payloadFieldError(kind, "schema", "must be "+strconv.Quote(SchemaOfficialQuoteL1))
	}
	currency, scale, known := marketMoney(quote.Market)
	if !known {
		return payloadFieldError(kind, "market", `must be "KR" or "US"`)
	}
	if quote.Currency != currency {
		return payloadFieldError(kind, "currency", "must be "+currency+" in market "+quote.Market)
	}
	if quote.PriceScale != scale {
		return payloadFieldError(kind, "price_scale", "must be "+formatUint(scale)+" for "+currency)
	}
	if !canonicalStoredSymbolText(quote.Symbol) {
		return payloadFieldError(kind, "symbol", "must be a canonical upper-case symbol without the ':' record separator")
	}
	for _, minor := range []struct {
		field string
		value uint64
	}{{"bid_minor", quote.BidMinor}, {"ask_minor", quote.AskMinor}, {"last_minor", quote.LastMinor}} {
		if minor.value == 0 {
			return payloadFieldError(kind, minor.field, "must be at least 1")
		}
	}
	if quote.BidMinor > quote.AskMinor {
		return payloadFieldError(kind, "bid_minor", "must not be above ask_minor")
	}
	for _, pair := range []struct {
		field, rawField, raw string
		declared             uint64
	}{
		{"bid_minor", "raw.bid", quote.Raw.Bid, quote.BidMinor},
		{"ask_minor", "raw.ask", quote.Raw.Ask, quote.AskMinor},
		{"last_minor", "raw.last", quote.Raw.Last, quote.LastMinor},
	} {
		if err := checkRawMinor(kind, pair.field, pair.rawField, pair.raw, scale, pair.declared); err != nil {
			return err
		}
	}
	if quote.SourceObservedAtMS == 0 {
		return payloadFieldError(kind, "source_observed_at_ms", "must be a positive unix millisecond")
	}
	if quote.ReceivedAtMS < quote.SourceObservedAtMS {
		return payloadFieldError(kind, "received_at_ms", "must not be before source_observed_at_ms")
	}
	if !canonicalSHA256Digest(quote.SourceResponseDigest) {
		return payloadFieldError(kind, "source_response_digest", "must be sha256: followed by 64 lowercase hex digits")
	}
	if quote.Revision == 0 {
		return payloadFieldError(kind, "revision", "must be at least 1")
	}
	return nil
}

// ---- 읽기 도구 ----

// payloadReader는 이름표가 붙은 JSON 물건을 하나씩 꺼내면서 "꺼낸 이름"을 적어 둔다.
// 마지막에 적히지 않은 이름이 남아 있으면 그것이 모르는 필드다.
type payloadReader struct {
	kind   EvidenceKind
	prefix string
	object map[string]any
	used   map[string]struct{}
	fault  *error
}

func newPayloadReader(kind EvidenceKind, object map[string]any) *payloadReader {
	var fault error
	return &payloadReader{kind: kind, object: object, used: make(map[string]struct{}), fault: &fault}
}

func (r *payloadReader) fail(field, detail string) {
	if *r.fault == nil {
		*r.fault = payloadFieldError(r.kind, r.prefix+field, detail)
	}
}

func (r *payloadReader) take(name string) (any, bool) {
	value, exists := r.object[name]
	if !exists {
		r.fail(name, "is required")
		return nil, false
	}
	r.used[name] = struct{}{}
	return value, true
}

func (r *payloadReader) text(name string) string {
	value, ok := r.take(name)
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		r.fail(name, "must be a JSON string")
		return ""
	}
	return text
}

func (r *payloadReader) flag(name string) bool {
	value, ok := r.take(name)
	if !ok {
		return false
	}
	flag, ok := value.(bool)
	if !ok {
		r.fail(name, "must be a JSON boolean")
		return false
	}
	return flag
}

func (r *payloadReader) integer(name string) uint64 {
	value, ok := r.take(name)
	if !ok {
		return 0
	}
	number, ok := value.(json.Number)
	if !ok {
		r.fail(name, "must be an integer JSON number")
		return 0
	}
	parsed, err := canonicalIntegerValue(string(number))
	if err != nil {
		r.fail(name, "must be an integer JSON number: "+err.Error())
		return 0
	}
	return parsed
}

func (r *payloadReader) child(name string) *payloadReader {
	nested := map[string]any{}
	if value, ok := r.take(name); ok {
		object, isObject := value.(map[string]any)
		if !isObject {
			r.fail(name, "must be a JSON object")
		} else {
			nested = object
		}
	}
	return &payloadReader{kind: r.kind, prefix: name + ".", object: nested, used: make(map[string]struct{}), fault: r.fault}
}

func (r *payloadReader) done() error {
	var unknown []string
	for name := range r.object {
		if _, ok := r.used[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 && *r.fault == nil {
		sort.Strings(unknown)
		*r.fault = fmt.Errorf("%s payload has unknown field %q", r.kind, r.prefix+unknown[0])
	}
	return *r.fault
}

func payloadFieldError(kind EvidenceKind, field, detail string) error {
	return fmt.Errorf("%s payload field %q %s", kind, field, detail)
}

// decodeCanonicalObject는 먼저 canonicalJSON을 통과시킨다.
// 그래야 중복 열쇠나 뒤에 붙은 값 같은, 정규형이 아닌 입력이 조용히 지나가지 못한다.
func decodeCanonicalObject(input []byte) (map[string]any, error) {
	canonical, err := canonicalJSON(input)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, errors.New("top-level JSON object is required")
	}
	return object, nil
}

// canonicalIntegerValue는 정규화된 JSON 숫자 글자를 정확한 정수로 되읽는다.
// canonicalJSON이 60000을 "6e4"로, 1.50을 "15e-1"로 바꾸기 때문에 글자 모양만 보면 안 된다.
// 음수 부호·소수점·음수 지수가 있으면 정수가 아니므로 거절한다.
func canonicalIntegerValue(text string) (uint64, error) {
	digits, exponent := text, 0
	if index := strings.IndexAny(text, "eE"); index >= 0 {
		digits = text[:index]
		parsed, err := strconv.Atoi(text[index+1:])
		if err != nil {
			return 0, errors.New("exponent is not an integer")
		}
		if parsed < 0 {
			return 0, errors.New("value is not a whole number")
		}
		if parsed > 20 {
			return 0, errors.New("value does not fit in 64 bits")
		}
		exponent = parsed
	}
	if strings.ContainsRune(digits, '.') {
		return 0, errors.New("value is not a whole number")
	}
	value, err := strconv.ParseUint(digits, 10, 64)
	if err != nil {
		return 0, errors.New("value is not a non-negative whole number")
	}
	for step := 0; step < exponent; step++ {
		if value > maxUint64/10 {
			return 0, errors.New("value does not fit in 64 bits")
		}
		value *= 10
	}
	return value, nil
}

// minorFromRawDecimal은 브로커가 준 소수 문자열을 정수 minor 단위로 다시 계산한다.
// 곱셈과 덧셈만 쓰므로 실수 오차가 끼어들 자리가 없다.
func minorFromRawDecimal(raw string, scale uint64) (uint64, error) {
	if raw == "" {
		return 0, errors.New("must be a plain decimal string")
	}
	integerPart, fractionPart := raw, ""
	if index := strings.IndexByte(raw, '.'); index >= 0 {
		integerPart, fractionPart = raw[:index], raw[index+1:]
	}
	if integerPart == "" || !onlyDigits(integerPart) || !onlyDigits(fractionPart) ||
		(strings.IndexByte(raw, '.') >= 0 && fractionPart == "") {
		return 0, errors.New("must be plain decimal digits with at most one fraction point and no sign, exponent or space")
	}
	// 정수부에 쓸데없는 앞자리 0이 붙으면 같은 값이 여러 글자로 저장되어 지문이 갈린다.
	if len(integerPart) > 1 && integerPart[0] == '0' {
		return 0, errors.New("must not carry a leading zero in its integer part")
	}
	if uint64(len(fractionPart)) > scale {
		return 0, errors.New("has more fraction digits than the declared scale " + formatUint(scale))
	}
	value := uint64(0)
	for index := 0; index < len(integerPart); index++ {
		next, ok := appendDecimalDigit(value, integerPart[index]-'0')
		if !ok {
			return 0, errors.New("does not fit in 64 bits")
		}
		value = next
	}
	for position := uint64(0); position < scale; position++ {
		digit := byte(0)
		if position < uint64(len(fractionPart)) {
			digit = fractionPart[position] - '0'
		}
		next, ok := appendDecimalDigit(value, digit)
		if !ok {
			return 0, errors.New("does not fit in 64 bits")
		}
		value = next
	}
	return value, nil
}

func checkRawMinor(kind EvidenceKind, field, rawField, raw string, scale, declared uint64) error {
	value, err := minorFromRawDecimal(raw, scale)
	if err != nil {
		return payloadFieldError(kind, rawField, err.Error())
	}
	if value != declared {
		return payloadFieldError(kind, field, "must equal "+rawField+" recomputed at scale "+formatUint(scale))
	}
	return nil
}

func appendDecimalDigit(value uint64, digit byte) (uint64, bool) {
	if value > maxUint64/10 {
		return 0, false
	}
	value *= 10
	if value > maxUint64-uint64(digit) {
		return 0, false
	}
	return value + uint64(digit), true
}

func onlyDigits(value string) bool {
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

// ---- 작은 공통 규칙 ----

// marketMoney는 시장마다 정해진 돈 단위와 소수 자릿수를 알려준다.
// 원화는 소수점이 없고(0), 달러는 소수 넷째 자리까지(4) 쓴다.
// 달러가 4인 이유: M-B 실측에서 미국 가격 문자열이 소수 1~4자리로 왔다.
// 자릿수가 적게 온 값("231.1")은 그대로 받아 뒤를 0으로 채우고,
// 선언한 자릿수보다 많이 온 값("231.43501")은 반올림하지 않고 거절한다.
func marketMoney(code string) (currency string, scale uint64, known bool) {
	switch code {
	case "KR":
		return "KRW", 0, true
	case "US":
		return "USD", 4, true
	default:
		return "", 0, false
	}
}

func marketCalendar(code string) (string, bool) {
	switch code {
	case "KR":
		return "KRX", true
	case "US":
		return "US", true
	default:
		return "", false
	}
}

// sessionDateFor는 "<달력>:<YYYY-MM-DD>" 모양의 세션 이름을 검사하고 날짜만 돌려준다.
func sessionDateFor(code, sessionID string) (string, error) {
	calendar, known := marketCalendar(code)
	if !known {
		return "", errors.New(`market must be "KR" or "US"`)
	}
	prefix := calendar + ":"
	if !strings.HasPrefix(sessionID, prefix) {
		return "", errors.New("must look like " + strconv.Quote(prefix+"YYYY-MM-DD"))
	}
	date := sessionID[len(prefix):]
	if len(date) != len("2006-01-02") {
		return "", errors.New("must look like " + strconv.Quote(prefix+"YYYY-MM-DD"))
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", errors.New("must carry a YYYY-MM-DD session date")
	}
	return date, nil
}

func marketAndCode(market marketclock.Market) (marketclock.Market, string, error) {
	normalized, err := marketclock.ParseMarket(string(market))
	if err != nil {
		return "", "", invalid(RefusalIdentityMismatch, "market", "only kr and us are supported")
	}
	return normalized, strings.ToUpper(string(normalized)), nil
}

func checkMarketCurrency(code, currency string) error {
	want, _, known := marketMoney(code)
	if !known {
		return invalid(RefusalIdentityMismatch, "market", "only kr and us are supported")
	}
	if strings.ToUpper(strings.TrimSpace(currency)) != want {
		return invalid(RefusalCurrencyUnresolved, "currency", "market "+code+" settles in "+want)
	}
	return nil
}

// canonicalSymbolText는 종목 코드에만 쓰는 더 엄한 규칙이다.
// ':'는 기록 번호를 나누는 글자라서 종목 안에 들어오면 신원이 애매해진다.
func canonicalSymbolText(value string) bool {
	return canonicalPayloadText(value) && !strings.ContainsRune(value, ':')
}

// canonicalStoredSymbolText는 저장되는 본문용이다. 머리말은 대문자로 바뀌어 저장되므로
// 본문도 처음부터 대문자여야 쓸 때 통과한 줄이 읽을 때 영영 거절되는 일이 없다.
func canonicalStoredSymbolText(value string) bool {
	return canonicalSymbolText(value) && strings.ToUpper(value) == value
}

func canonicalPayloadText(value string) bool {
	return value != "" && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsRune(value, '\x00')
}

func canonicalSHA256Digest(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	hexDigits := value[len(prefix):]
	if len(hexDigits) != 64 {
		return false
	}
	for index := 0; index < len(hexDigits); index++ {
		character := hexDigits[index]
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

// closedBar1mRecordID는 봉 하나의 신원 문자열이다. 머리말을 만들 때와 읽어서 확인할 때
// 반드시 같은 글자가 나와야 하므로 만드는 곳을 한 군데로 모았다.
// 종목에 ':'를 금지했기 때문에 이 문자열은 원래 값들로 다시 쪼갤 수 있다.
func closedBar1mRecordID(code, symbol, sessionID string, intervalMS, openAtMS uint64) string {
	return strings.Join([]string{code, symbol, sessionID, formatUint(intervalMS), formatUint(openAtMS)}, ":")
}

func revisionIdentityFor(revision uint64) string { return "r" + formatUint(revision) }

func supersededRevisionIdentity(revision uint64) string {
	if revision <= 1 {
		return ""
	}
	return revisionIdentityFor(revision - 1)
}

func evidenceIDFor(sourceRecordID string, revision uint64) string {
	return sourceRecordID + ":" + revisionIdentityFor(revision)
}

func formatUint(value uint64) string { return strconv.FormatUint(value, 10) }
