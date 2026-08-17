package strategyevidence

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 아래 두 값은 breakout 코드를 넣기 *전*에 실제로 측정해서 박아 둔 값이다.
// 새 종류(kind)를 추가한 뒤에도 옛 종류의 바이트와 지문이 한 글자도 변하지 않았음을 증명한다.
const (
	legacyParticipationCanonicalJSON = `{"blocked":true,"code":"x","score_ppm":25e1,"value":1}`
	legacyParticipationDigest        = "7e30f2af1ce6969ebd444a015a5d3ac17245d29ed7c7deb8f7012aa9ffbea851"
)

const (
	usTestSessionID       = "US:2026-08-14"
	usTestCalendarVersion = "us-calendar-v1"
	krTestSessionID       = "KRX:2026-08-18"
	krTestCalendarVersion = "krx-calendar-v1"
)

// 시험용 응답 지문. 진짜 응답에서 온 값이 아니라 모양만 맞춘 손으로 쓴 값이다.
var testResponseDigest = "sha256:" + strings.Repeat("ab", 32)

// 미국 정규장 첫 1분봉(09:30 ET = 13:30 UTC)과 한국 정규장 첫 1분봉(09:00 KST = 00:00 UTC).
var (
	usTestBarOpenAt = time.Date(2026, 8, 14, 13, 30, 0, 0, time.UTC)
	krTestBarOpenAt = time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
)

func TestLegacyKindCanonicalBytesAndDigestAreUnchanged(t *testing.T) {
	t.Parallel()
	header := validHeader(marketclock.MarketUS, KindUSParticipation, "legacy-guard", "rev-1")
	envelope := mustEnvelope(t, header, `{"value": 1, "code": "x", "blocked": true, "score_ppm": 250}`)
	if got := string(envelope.CanonicalPayload()); got != legacyParticipationCanonicalJSON {
		t.Fatalf("legacy canonical payload changed:\n got  %s\n want %s", got, legacyParticipationCanonicalJSON)
	}
	if got := envelope.PayloadDigest(); got != legacyParticipationDigest {
		t.Fatalf("legacy payload digest changed: got %s want %s", got, legacyParticipationDigest)
	}
}

func TestKindSupportsMarketCoversBreakoutKindsInKRAndUS(t *testing.T) {
	t.Parallel()
	for _, kind := range []EvidenceKind{KindOfficialClosedBar1m, KindOfficialQuoteL1} {
		for _, market := range []marketclock.Market{marketclock.MarketKR, marketclock.MarketUS} {
			if !kindSupportsMarket(kind, market) {
				t.Fatalf("kindSupportsMarket(%s, %s) = false, want true", kind, market)
			}
		}
	}
	// 기존 종류의 시장 제한이 그대로인지도 같이 지킨다.
	for _, tt := range []struct {
		kind   EvidenceKind
		market marketclock.Market
		want   bool
	}{
		{KindKRNetFlow, marketclock.MarketKR, true},
		{KindKRNetFlow, marketclock.MarketUS, false},
		{KindUSParticipation, marketclock.MarketUS, true},
		{KindUSParticipation, marketclock.MarketKR, false},
		{"official_closed_bar_5m", marketclock.MarketUS, false},
		{"", marketclock.MarketUS, false},
	} {
		if got := kindSupportsMarket(tt.kind, tt.market); got != tt.want {
			t.Fatalf("kindSupportsMarket(%q, %s) = %v, want %v", tt.kind, tt.market, got, tt.want)
		}
	}
}

func TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars(t *testing.T) {
	t.Parallel()
	usEnvelope, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), nil))
	if err != nil {
		t.Fatalf("US closed bar refused: %v", err)
	}
	decoded, err := DecodeClosedBar1mPayload(usEnvelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("US closed bar replay refused: %v", err)
	}
	if decoded.CloseMinor != 2316500 || decoded.PriceScale != 4 || decoded.Currency != "USD" || decoded.Raw.Close != "231.65" {
		t.Fatalf("US closed bar decoded wrongly: %+v", decoded)
	}
	// 소수 4자리·1자리·2자리가 모두 같은 scale 4에서 정확히 정수로 환산되어야 한다.
	if decoded.OpenMinor != 2314350 || decoded.LowMinor != 2311000 || decoded.HighMinor != 2318000 {
		t.Fatalf("US closed bar minors recomputed wrongly: %+v", decoded)
	}
	if decoded.IntervalMS != 60000 || decoded.OpenAtMS != uint64(usTestBarOpenAt.UnixMilli()) {
		t.Fatalf("US closed bar identity decoded wrongly: %+v", decoded)
	}

	krHeader, err := newClosedBar1mHeader(marketclock.MarketKR, "005930", krTestSessionID, krTestCalendarVersion,
		krTestBarOpenAt, 1, krTestBarOpenAt.Add(2*time.Minute), "KRW")
	if err != nil {
		t.Fatalf("KR bar header refused: %v", err)
	}
	krEnvelope, err := NewEnvelope(krHeader, barPayloadBytes(t, krTestBarFields(uint64(krTestBarOpenAt.UnixMilli())), nil))
	if err != nil {
		t.Fatalf("KR closed bar refused: %v", err)
	}
	krDecoded, err := DecodeClosedBar1mPayload(krEnvelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("KR closed bar replay refused: %v", err)
	}
	if krDecoded.PriceScale != 0 || krDecoded.Currency != "KRW" || krDecoded.CloseMinor != 71300 {
		t.Fatalf("KR closed bar decoded wrongly: %+v", krDecoded)
	}
}

func TestClosedBarRejectsUnknownField(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "top level", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["adjusted"] = false
	}), "unknown field")
	assertBarPayloadRefused(t, "inside raw", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["vwap"] = "231.65"
	}), "unknown field")
}

func TestClosedBarRejectsUnknownEnumValues(t *testing.T) {
	t.Parallel()
	// 열거값마다 낯선 값 셋 이상을 넣는다. 빈 문자열과 대소문자만 다른 값이 반드시 들어간다.
	for _, tt := range []struct {
		field  string
		values []any
	}{
		{"bar_label", []any{"close_at", "", "OPEN_AT", "Open_At", "open_at ", "opened_at"}},
		{"finality", []any{"unknown", "", "SUCCESSOR_OBSERVED", "Successor_Observed", "successor-observed", "final"}},
		{"schema", []any{"official_closed_bar_1m:v2", "", "OFFICIAL_CLOSED_BAR_1M:V1", "official_closed_bar_1m", "official_quote_l1:v1"}},
		{"currency", []any{"EUR", "", "usd", "Usd", "KRW", "USDT"}},
		{"market", []any{"JP", "", "us", "Us", "USA", "KR"}},
		{"session_id", []any{"KRX:2026-08-14", "", "us:2026-08-14", "Us:2026-08-14", "US-2026-08-14", "2026-08-14"}},
	} {
		tt := tt
		t.Run(tt.field, func(t *testing.T) {
			t.Parallel()
			for _, value := range tt.values {
				assertBarPayloadRefused(t, tt.field, barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
					fields[tt.field] = value
				}), tt.field)
			}
		})
	}
}

func TestQuoteL1RejectsUnknownEnumValues(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		field  string
		values []any
	}{
		{"schema", []any{"official_quote_l1:v2", "", "OFFICIAL_QUOTE_L1:V1", "official_closed_bar_1m:v1"}},
		{"currency", []any{"EUR", "", "usd", "Usd", "KRW"}},
		{"market", []any{"JP", "", "us", "Us", "KR"}},
	} {
		tt := tt
		t.Run(tt.field, func(t *testing.T) {
			t.Parallel()
			for _, value := range tt.values {
				_, err := NewEnvelope(usTestQuoteHeader(t, 1), quotePayloadBytes(t, usTestQuoteFields(), func(fields map[string]any) {
					fields[tt.field] = value
				}))
				var refusal *ValidationError
				if !errors.As(err, &refusal) || !strings.Contains(refusal.Detail, tt.field) {
					t.Fatalf("quote %s=%v: want a refusal naming the field, got %v", tt.field, value, err)
				}
			}
		})
	}
}

func TestClosedBarRejectsDecimalPriceNumber(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "decimal close", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["close_minor"] = json.RawMessage("231.65")
	}), "close_minor")
	assertBarPayloadRefused(t, "decimal volume", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["volume"] = json.RawMessage("12345.5")
	}), "volume")
	assertBarPayloadRefused(t, "negative minor", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["low_minor"] = json.RawMessage("-2311000")
	}), "low_minor")
	assertBarPayloadRefused(t, "exponent minor", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["high_minor"] = json.RawMessage("1.90805e4")
	}), "high_minor")
}

// canonicalJSON이 먼저 숫자를 정규형으로 바꾸므로(60000 → "6e4", 1.50 → "15e-1")
// "정수인가"는 글자 모양이 아니라 값으로 판단한다. 값이 정확히 정수면 받고, 아니면 거절한다.
func TestClosedBarIntegerRuleIsAboutTheValueNotTheSpelling(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	accepted, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["volume"] = json.RawMessage("1.2345e4")
	}))
	if err != nil {
		t.Fatalf("exactly integral volume written with an exponent was refused: %v", err)
	}
	decoded, err := DecodeClosedBar1mPayload(accepted.CanonicalPayload())
	if err != nil || decoded.Volume != 12345 {
		t.Fatalf("volume = %d, err = %v, want 12345", decoded.Volume, err)
	}
	assertBarPayloadRefused(t, "fractional volume written with an exponent",
		barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
			fields["volume"] = json.RawMessage("1.23455e4")
		}), "volume")
}

func TestClosedBarRejectsBareNumberRawPriceAndStringMinor(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "bare number raw price", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["close"] = json.RawMessage("231.65")
	}), "close")
	assertBarPayloadRefused(t, "string minor", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["open_minor"] = "2314350"
	}), "open_minor")
	assertBarPayloadRefused(t, "string interval", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["interval_ms"] = "60000"
	}), "interval_ms")
	assertBarPayloadRefused(t, "number symbol", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["symbol"] = json.RawMessage("42")
	}), "symbol")
	assertBarPayloadRefused(t, "string closed flag", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["closed"] = "true"
	}), "closed")
}

func TestClosedBarRejectsMinorThatDisagreesWithRawDecimal(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "close mismatch", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["close_minor"] = uint64(2316501)
	}), "close_minor")
	assertBarPayloadRefused(t, "volume mismatch", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["volume"] = "12346"
	}), "volume")
}

func TestClosedBarRejectsOverPreciseRawForTheDeclaredScale(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "usd five decimals", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["close"] = "231.43501"
	}), "close")
	assertBarPayloadRefused(t, "krw fraction", barPayloadBytes(t, krTestBarFields(uint64(krTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["close"] = "71300.5"
	}), "close")
	for _, declared := range []uint64{0, 2, 3, 5} {
		declared := declared
		assertBarPayloadRefused(t, "declared scale disagrees with currency", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
			fields["price_scale"] = declared
		}), "price_scale")
	}
}

func TestClosedBarRejectsSignedExponentOrPaddedRawDecimal(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"+231.65", "-231.65", "2.3165e2", " 231.65", "231.65 ", "231,65", "231.", ".65", "", "23I.65"} {
		raw := raw
		t.Run(strings.ReplaceAll(raw, " ", "_"), func(t *testing.T) {
			t.Parallel()
			assertBarPayloadRefused(t, "raw close "+raw, barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
				fields["raw"].(map[string]any)["close"] = raw
			}), "close")
		})
	}
}

func TestClosedBarRejectsSecretLikeField(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "access token", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["access_token"] = "must-not-store"
	}), "secret-like")
	assertBarPayloadRefused(t, "nested credential", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["api_key"] = "must-not-store"
	}), "secret-like")
}

func TestClosedBarRejectsIntervalOtherThanOneMinute(t *testing.T) {
	t.Parallel()
	for _, interval := range []uint64{0, 1000, 30000, 60001, 300000} {
		interval := interval
		assertBarPayloadRefused(t, "interval", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
			fields["interval_ms"] = interval
		}), "interval_ms")
	}
}

func TestClosedBarRejectsOpenAtThatIsNotOnTheMinute(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	for _, openAt := range []uint64{base + 1, base + 59999, base - 30000, 0} {
		openAt := openAt
		assertBarPayloadRefused(t, "open_at_ms", barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
			fields["open_at_ms"] = openAt
		}), "open_at_ms")
	}
}

func TestClosedBarRejectsBarFromTheFuture(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	for _, observed := range []uint64{base, base - 60000, 0} {
		observed := observed
		assertBarPayloadRefused(t, "future bar", barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
			fields["source_observed_at_ms"] = observed
		}), "source_observed_at_ms")
	}
}

func TestClosedBarRejectsUnfinishedBar(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "closed false", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["closed"] = false
	}), "closed")
}

func TestClosedBarRejectsEveryMissingRequiredField(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	var names []string
	for name := range usTestBarFields(base) {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) != 23 {
		t.Fatalf("closed bar fixture has %d fields, want the 23 required ones: %v", len(names), names)
	}
	for _, name := range names {
		name := name
		assertBarPayloadRefused(t, "missing "+name, barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
			delete(fields, name)
		}), name)
	}
	for _, name := range []string{"open", "high", "low", "close", "volume"} {
		name := name
		assertBarPayloadRefused(t, "missing raw."+name, barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
			delete(fields["raw"].(map[string]any), name)
		}), name)
	}
	assertBarPayloadRefused(t, "raw is not an object", barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["raw"] = "231.65"
	}), "raw")
}

func TestClosedBarRejectsImpossiblePriceOrdering(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	for _, tt := range []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{"low above high", func(fields map[string]any) {
			fields["low_minor"], fields["raw"].(map[string]any)["low"] = uint64(2319000), "231.9"
		}, "low_minor"},
		{"close above high", func(fields map[string]any) {
			fields["close_minor"], fields["raw"].(map[string]any)["close"] = uint64(2318500), "231.85"
		}, "close_minor"},
		{"open below low", func(fields map[string]any) {
			fields["open_minor"], fields["raw"].(map[string]any)["open"] = uint64(2310000), "231.0"
		}, "open_minor"},
		{"zero low", func(fields map[string]any) {
			fields["low_minor"], fields["raw"].(map[string]any)["low"] = uint64(0), "0"
		}, "low_minor"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertBarPayloadRefused(t, tt.name, barPayloadBytes(t, usTestBarFields(base), tt.mutate), tt.want)
		})
	}
}

func TestClosedBarRejectsMalformedIdentityStrings(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	for _, tt := range []struct {
		name  string
		field string
		value any
	}{
		{"session calendar does not match market", "session_id", "KRX:2026-08-14"},
		{"session date is not a date", "session_id", "US:2026-08-32"},
		{"session identity has no calendar", "session_id", "2026-08-14"},
		{"symbol has surrounding space", "symbol", " AAPL"},
		{"symbol is empty", "symbol", ""},
		{"symbol carries NUL", "symbol", "AA\x00PL"},
		{"calendar version is empty", "calendar_version", ""},
		{"digest has no prefix", "source_response_digest", strings.Repeat("ab", 32)},
		{"digest is upper case", "source_response_digest", "sha256:" + strings.Repeat("AB", 32)},
		{"digest is short", "source_response_digest", "sha256:" + strings.Repeat("ab", 31)},
		{"revision is zero", "revision", uint64(0)},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertBarPayloadRefused(t, tt.name, barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
				fields[tt.field] = tt.value
			}), tt.field)
		})
	}
}

// 이 시험은 새 종류의 갈림길이 옛 종류의 얕은 타입 지도보다 *먼저* 온다는 순서를 지킨다.
// 갈림길이 뒤로 밀리면 "value" 같은 옛 이름은 옛 지도가 먼저 걸러서 다른 이유로 거절된다.
func TestClosedBarDispatchRunsBeforeTheLegacyTypeMap(t *testing.T) {
	t.Parallel()
	payload := barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["value"] = "not-a-number"
	})
	_, err := NewEnvelope(usTestBarHeader(t, 1), payload)
	var refusal *ValidationError
	if !errors.As(err, &refusal) {
		t.Fatalf("want a typed refusal, got %v", err)
	}
	if !strings.Contains(refusal.Detail, "unknown field") {
		t.Fatalf("refusal detail = %q, want the strict breakout decoder to report an unknown field", refusal.Detail)
	}
	if strings.Contains(refusal.Detail, "must be number") {
		t.Fatalf("legacy generic type map ran before the breakout dispatch: %q", refusal.Detail)
	}
}

func TestQuoteL1EnvelopeAcceptsCanonicalQuote(t *testing.T) {
	t.Parallel()
	header := usTestQuoteHeader(t, 1)
	envelope, err := NewEnvelope(header, quotePayloadBytes(t, usTestQuoteFields(), nil))
	if err != nil {
		t.Fatalf("quote refused: %v", err)
	}
	decoded, err := DecodeQuoteL1Payload(envelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("quote replay refused: %v", err)
	}
	if decoded.BidMinor != 2316000 || decoded.AskMinor != 2317000 || decoded.LastMinor != 2316500 || decoded.Currency != "USD" {
		t.Fatalf("quote decoded wrongly: %+v", decoded)
	}
	if decoded.PriceScale != 4 {
		t.Fatalf("quote price scale = %d, want 4", decoded.PriceScale)
	}
	if decoded.SourceObservedAtMS > decoded.ReceivedAtMS || decoded.Revision != 1 {
		t.Fatalf("quote clocks decoded wrongly: %+v", decoded)
	}
}

func TestQuoteL1RejectsEveryContractViolation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{"unknown field", func(fields map[string]any) { fields["spread_ppm"] = uint64(5) }, "unknown field"},
		{"unknown raw field", func(fields map[string]any) { fields["raw"].(map[string]any)["mid"] = "231.65" }, "unknown field"},
		{"bid above ask", func(fields map[string]any) {
			fields["bid_minor"], fields["raw"].(map[string]any)["bid"] = uint64(2317500), "231.75"
		}, "bid_minor"},
		{"zero last", func(fields map[string]any) {
			fields["last_minor"], fields["raw"].(map[string]any)["last"] = uint64(0), "0"
		}, "last_minor"},
		{"raw disagrees with minor", func(fields map[string]any) { fields["raw"].(map[string]any)["ask"] = "231.7001" }, "ask"},
		{"over precise raw", func(fields map[string]any) { fields["raw"].(map[string]any)["ask"] = "231.70001" }, "ask"},
		{"decimal minor", func(fields map[string]any) { fields["bid_minor"] = json.RawMessage("2316000.5") }, "bid_minor"},
		{"received before observed", func(fields map[string]any) {
			fields["received_at_ms"] = fields["source_observed_at_ms"].(uint64) - 1
		}, "received_at_ms"},
		{"zero observed clock", func(fields map[string]any) { fields["source_observed_at_ms"] = uint64(0) }, "source_observed_at_ms"},
		{"currency does not match market", func(fields map[string]any) { fields["currency"] = "KRW" }, "currency"},
		{"unknown schema", func(fields map[string]any) { fields["schema"] = "official_quote_l1:v2" }, "schema"},
		{"secret-like field", func(fields map[string]any) { fields["session_token"] = "x" }, "secret-like"},
		{"zero revision", func(fields map[string]any) { fields["revision"] = uint64(0) }, "revision"},
		{"missing digest", func(fields map[string]any) { delete(fields, "source_response_digest") }, "source_response_digest"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewEnvelope(usTestQuoteHeader(t, 1), quotePayloadBytes(t, usTestQuoteFields(), tt.mutate))
			var refusal *ValidationError
			if !errors.As(err, &refusal) || refusal.Code != RefusalPayloadInvalid {
				t.Fatalf("%s: want payload refusal, got %v", tt.name, err)
			}
			if !strings.Contains(refusal.Detail, tt.want) {
				t.Fatalf("%s: refusal detail = %q, want it to mention %q", tt.name, refusal.Detail, tt.want)
			}
		})
	}
}

func TestClosedBar1mHeaderCarriesDeterministicBarIdentity(t *testing.T) {
	t.Parallel()
	openAtMS := usTestBarOpenAt.UnixMilli()
	observedAt := usTestBarOpenAt.Add(2 * time.Minute)
	first, err := newClosedBar1mHeader(marketclock.MarketUS, "aapl", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, observedAt, "usd")
	if err != nil {
		t.Fatalf("header refused: %v", err)
	}
	again, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, observedAt, "USD")
	if err != nil {
		t.Fatalf("header refused: %v", err)
	}
	if first != again {
		t.Fatalf("closed bar header is not deterministic:\n%+v\n%+v", first, again)
	}
	wantRecord := "US:AAPL:US:2026-08-14:60000:" + formatUint(uint64(openAtMS))
	if first.SourceRecordID != wantRecord {
		t.Fatalf("source record id = %q, want %q", first.SourceRecordID, wantRecord)
	}
	if first.EvidenceID != wantRecord+":r1" {
		t.Fatalf("evidence id = %q, want %q", first.EvidenceID, wantRecord+":r1")
	}
	if first.RevisionIdentity != "r1" || first.SupersedesRevisionIdentity != "" {
		t.Fatalf("revision identity = %q / %q", first.RevisionIdentity, first.SupersedesRevisionIdentity)
	}
	if first.IssuerIdentity != "US:AAPL" || first.IssuerMappingVersion != "a112-bar-issuer-v1" {
		t.Fatalf("issuer = %q / %q", first.IssuerIdentity, first.IssuerMappingVersion)
	}
	if first.Kind != KindOfficialClosedBar1m || first.SchemaVersion != "official_closed_bar_1m:v1" || first.Authority != AuthorityTossOpenAPI {
		t.Fatalf("kind/schema/authority = %q / %q / %q", first.Kind, first.SchemaVersion, first.Authority)
	}
	if first.MarketEffectiveDate != "2026-08-14" || first.Currency != "USD" || first.Unit != "minor" {
		t.Fatalf("effective date/currency/unit = %q / %q / %q", first.MarketEffectiveDate, first.Currency, first.Unit)
	}
	if first.Availability != AvailabilityAvailable || first.Confidence != ConfidenceVerified {
		t.Fatalf("availability/confidence = %q / %q", first.Availability, first.Confidence)
	}
	if !first.SourceEventAt.Equal(usTestBarOpenAt) || !first.SourceAvailableAt.Equal(usTestBarOpenAt.Add(time.Minute)) {
		t.Fatalf("source clocks = %s / %s", first.SourceEventAt, first.SourceAvailableAt)
	}
	if !first.ObservedAt.Equal(observedAt) || !first.IngestedAt.Equal(observedAt) {
		t.Fatalf("observed/ingested = %s / %s", first.ObservedAt, first.IngestedAt)
	}
	corrected, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 2, observedAt, "USD")
	if err != nil {
		t.Fatalf("correction header refused: %v", err)
	}
	if corrected.RevisionIdentity != "r2" || corrected.SupersedesRevisionIdentity != "r1" || corrected.SourceRecordID != wantRecord {
		t.Fatalf("correction identity = %q / %q / %q", corrected.RevisionIdentity, corrected.SupersedesRevisionIdentity, corrected.SourceRecordID)
	}
	if corrected.EvidenceID == first.EvidenceID {
		t.Fatal("correction reused the evidence id of the revision it supersedes")
	}
}

func TestClosedBar1mHeaderRefusesInconsistentInput(t *testing.T) {
	t.Parallel()
	observedAt := usTestBarOpenAt.Add(2 * time.Minute)
	for _, tt := range []struct {
		name string
		call func() (Header, error)
	}{
		{"unknown market", func() (Header, error) {
			return newClosedBar1mHeader("jp", "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, observedAt, "USD")
		}},
		{"currency does not match market", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, observedAt, "KRW")
		}},
		{"session calendar does not match market", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", krTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, observedAt, "USD")
		}},
		{"revision zero", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 0, observedAt, "USD")
		}},
		{"empty symbol", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "  ", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, observedAt, "USD")
		}},
		{"empty calendar version", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, "", usTestBarOpenAt, 1, observedAt, "USD")
		}},
		{"open_at is not on the minute", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt.Add(time.Second), 1, observedAt, "USD")
		}},
		{"observed before the bar closed", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, usTestBarOpenAt.Add(30*time.Second), "USD")
		}},
		{"zero observed clock", func() (Header, error) {
			return newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion, usTestBarOpenAt, 1, time.Time{}, "USD")
		}},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			header, err := tt.call()
			var refusal *ValidationError
			if !errors.As(err, &refusal) {
				t.Fatalf("want a typed refusal, got header %+v err %v", header, err)
			}
		})
	}
}

func TestQuoteL1HeaderCarriesObservedInstantIdentity(t *testing.T) {
	t.Parallel()
	observedAt := time.Date(2026, 8, 14, 14, 0, 0, 0, time.UTC)
	receivedAt := observedAt.Add(250 * time.Millisecond)
	header, err := newQuoteL1Header(marketclock.MarketUS, "aapl", observedAt, 1, receivedAt, "USD")
	if err != nil {
		t.Fatalf("quote header refused: %v", err)
	}
	wantRecord := "US:AAPL:quote_l1:" + formatUint(uint64(observedAt.UnixMilli()))
	if header.SourceRecordID != wantRecord || header.EvidenceID != wantRecord+":r1" {
		t.Fatalf("quote identity = %q / %q", header.SourceRecordID, header.EvidenceID)
	}
	if header.Kind != KindOfficialQuoteL1 || header.SchemaVersion != "official_quote_l1:v1" {
		t.Fatalf("quote kind/schema = %q / %q", header.Kind, header.SchemaVersion)
	}
	if !header.SourceEventAt.Equal(observedAt) || !header.SourceAvailableAt.Equal(observedAt) || !header.ObservedAt.Equal(receivedAt) {
		t.Fatalf("quote clocks = %s / %s / %s", header.SourceEventAt, header.SourceAvailableAt, header.ObservedAt)
	}
	if header.MarketEffectiveDate != "2026-08-14" || header.IssuerIdentity != "US:AAPL" {
		t.Fatalf("quote effective date/issuer = %q / %q", header.MarketEffectiveDate, header.IssuerIdentity)
	}
	if _, err := newQuoteL1Header(marketclock.MarketUS, "AAPL", observedAt, 1, observedAt.Add(-time.Millisecond), "USD"); err == nil {
		t.Fatal("quote received before it was observed was accepted")
	}
}

// 결합 생성자는 머리말과 본문을 한 곳에서 만들므로 둘이 어긋날 수 없다.
func TestNewClosedBar1mEnvelopeDerivesScaleMinorsAndIdentityTogether(t *testing.T) {
	t.Parallel()
	envelope, err := NewClosedBar1mEnvelope(usTestBarInput())
	if err != nil {
		t.Fatalf("constructor refused a good bar: %v", err)
	}
	header := envelope.Header()
	wantHeader := usTestBarHeader(t, 1)
	if header != wantHeader {
		t.Fatalf("constructor header differs from the header helper:\n%+v\n%+v", header, wantHeader)
	}
	bar, err := DecodeClosedBar1mPayload(envelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("constructor built a payload its own decoder refuses: %v", err)
	}
	if bar.PriceScale != 4 || bar.Currency != "USD" {
		t.Fatalf("scale/currency were not derived from the market: %+v", bar)
	}
	if bar.OpenMinor != 2314350 || bar.HighMinor != 2318000 || bar.LowMinor != 2311000 || bar.CloseMinor != 2316500 || bar.Volume != 12345 {
		t.Fatalf("minors were not recomputed from the raw strings: %+v", bar)
	}
	if bar.Raw != (RawClosedBar1m{Open: "231.4350", High: "231.8000", Low: "231.1", Close: "231.65", Volume: "12345"}) {
		t.Fatalf("raw strings were not preserved byte for byte: %+v", bar.Raw)
	}
	if bar.Market != "US" || bar.Symbol != header.Symbol || bar.SessionID != usTestSessionID {
		t.Fatalf("payload scope disagrees with the header: %+v", bar)
	}
	if bar.OpenAtMS != uint64(header.SourceEventAt.UnixMilli()) || bar.SourceObservedAtMS != uint64(header.ObservedAt.UnixMilli()) {
		t.Fatalf("payload clocks disagree with the header: %+v", bar)
	}
	if bar.Closed != true || bar.Finality != "successor_observed" || bar.BarLabel != "open_at" || bar.IntervalMS != 60000 {
		t.Fatalf("constructor did not pin the closed-bar constants: %+v", bar)
	}
	if bar.Revision != 1 || bar.SourceResponseDigest != testResponseDigest {
		t.Fatalf("revision/digest were not carried: %+v", bar)
	}
}

func TestNewClosedBar1mEnvelopeRefusesInconsistentInput(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		mutate func(*ClosedBar1mInput)
	}{
		{"over precise raw close", func(input *ClosedBar1mInput) { input.Raw.Close = "231.65001" }},
		{"fractional raw volume", func(input *ClosedBar1mInput) { input.Raw.Volume = "12345.5" }},
		{"signed raw open", func(input *ClosedBar1mInput) { input.Raw.Open = "+231.4350" }},
		{"currency does not match market", func(input *ClosedBar1mInput) { input.Currency = "KRW" }},
		{"session calendar does not match market", func(input *ClosedBar1mInput) { input.SessionID = krTestSessionID }},
		{"unknown market", func(input *ClosedBar1mInput) { input.Market = "jp" }},
		{"revision zero", func(input *ClosedBar1mInput) { input.Revision = 0 }},
		{"open_at is not on the minute", func(input *ClosedBar1mInput) { input.OpenAt = input.OpenAt.Add(time.Second) }},
		{"observed at the bar open", func(input *ClosedBar1mInput) { input.ObservedAt = input.OpenAt }},
		{"observed before the bar open", func(input *ClosedBar1mInput) { input.ObservedAt = input.OpenAt.Add(-time.Minute) }},
		{"low above high", func(input *ClosedBar1mInput) { input.Raw.Low = "231.9" }},
		{"malformed digest", func(input *ClosedBar1mInput) { input.SourceResponseDigest = "sha256:nothex" }},
		{"empty symbol", func(input *ClosedBar1mInput) { input.Symbol = " " }},
		{"empty calendar version", func(input *ClosedBar1mInput) { input.CalendarVersion = "" }},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := usTestBarInput()
			tt.mutate(&input)
			envelope, err := NewClosedBar1mEnvelope(input)
			var refusal *ValidationError
			if !errors.As(err, &refusal) {
				t.Fatalf("%s: want a typed refusal, got %v", tt.name, err)
			}
			if envelope.PayloadDigest() != "" {
				t.Fatalf("%s: refusal returned an envelope", tt.name)
			}
		})
	}
}

func TestNewQuoteL1EnvelopeDerivesScaleAndMinorsTogether(t *testing.T) {
	t.Parallel()
	envelope, err := NewQuoteL1Envelope(usTestQuoteInput())
	if err != nil {
		t.Fatalf("constructor refused a good quote: %v", err)
	}
	header := envelope.Header()
	if header != usTestQuoteHeader(t, 1) {
		t.Fatalf("constructor header differs from the header helper: %+v", header)
	}
	quote, err := DecodeQuoteL1Payload(envelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("constructor built a payload its own decoder refuses: %v", err)
	}
	if quote.PriceScale != 4 || quote.BidMinor != 2316000 || quote.AskMinor != 2317000 || quote.LastMinor != 2316500 {
		t.Fatalf("quote minors were not recomputed at scale 4: %+v", quote)
	}
	if quote.SourceObservedAtMS != uint64(header.SourceEventAt.UnixMilli()) || quote.ReceivedAtMS != uint64(header.ObservedAt.UnixMilli()) {
		t.Fatalf("quote clocks disagree with the header: %+v", quote)
	}
	for _, tt := range []struct {
		name   string
		mutate func(*QuoteL1Input)
	}{
		{"over precise raw ask", func(input *QuoteL1Input) { input.Raw.Ask = "231.70001" }},
		{"bid above ask", func(input *QuoteL1Input) { input.Raw.Bid = "231.75" }},
		{"currency does not match market", func(input *QuoteL1Input) { input.Currency = "KRW" }},
		{"revision zero", func(input *QuoteL1Input) { input.Revision = 0 }},
		{"received before observed", func(input *QuoteL1Input) { input.ReceivedAt = input.SourceObservedAt.Add(-time.Millisecond) }},
		{"malformed digest", func(input *QuoteL1Input) { input.SourceResponseDigest = "nothex" }},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := usTestQuoteInput()
			tt.mutate(&input)
			if _, err := NewQuoteL1Envelope(input); err == nil {
				t.Fatalf("%s: constructor accepted an inconsistent quote", tt.name)
			}
		})
	}
}

// ---- 시험용 고정 자료 (손으로 쓴 모양만 맞춘 값; 실제 응답 본문이 아니다) ----

func usTestBarInput() ClosedBar1mInput {
	return ClosedBar1mInput{
		Market:               marketclock.MarketUS,
		Symbol:               "AAPL",
		SessionID:            usTestSessionID,
		CalendarVersion:      usTestCalendarVersion,
		OpenAt:               usTestBarOpenAt,
		Revision:             1,
		ObservedAt:           usTestBarOpenAt.Add(2 * time.Minute),
		Currency:             "USD",
		Raw:                  RawClosedBar1m{Open: "231.4350", High: "231.8000", Low: "231.1", Close: "231.65", Volume: "12345"},
		SuccessorOpenAt:      usTestBarOpenAt.Add(time.Minute),
		RegularSession:       true,
		SourceResponseDigest: testResponseDigest,
	}
}

func usTestQuoteInput() QuoteL1Input {
	observed := time.Date(2026, 8, 14, 14, 0, 0, 0, time.UTC)
	return QuoteL1Input{
		Market:               marketclock.MarketUS,
		Symbol:               "AAPL",
		SourceObservedAt:     observed,
		ReceivedAt:           observed.Add(250 * time.Millisecond),
		Revision:             1,
		Currency:             "USD",
		Raw:                  RawQuoteL1{Bid: "231.6000", Ask: "231.7000", Last: "231.65"},
		SourceResponseDigest: testResponseDigest,
	}
}

func usTestBarFields(openAtMS uint64) map[string]any {
	return map[string]any{
		"schema":           "official_closed_bar_1m:v1",
		"market":           "US",
		"symbol":           "AAPL",
		"session_id":       usTestSessionID,
		"calendar_version": usTestCalendarVersion,
		"interval_ms":      uint64(60000),
		"bar_label":        "open_at",
		"open_at_ms":       openAtMS,
		"finality":         "successor_observed",
		"closed":           true,
		"regular_session":  true,
		"currency":         "USD",
		"price_scale":      uint64(4),
		"open_minor":       uint64(2314350),
		"high_minor":       uint64(2318000),
		"low_minor":        uint64(2311000),
		"close_minor":      uint64(2316500),
		"volume":           uint64(12345),
		// 자릿수가 선언한 4보다 적게 온 값("231.1", "231.65")도 그대로 받아야 한다.
		"raw":                    map[string]any{"open": "231.4350", "high": "231.8000", "low": "231.1", "close": "231.65", "volume": "12345"},
		"revision":               uint64(1),
		"successor_open_at_ms":   openAtMS + 60000,
		"source_observed_at_ms":  openAtMS + 120000,
		"source_response_digest": testResponseDigest,
	}
}

func krTestBarFields(openAtMS uint64) map[string]any {
	return map[string]any{
		"schema":                 "official_closed_bar_1m:v1",
		"market":                 "KR",
		"symbol":                 "005930",
		"session_id":             krTestSessionID,
		"calendar_version":       krTestCalendarVersion,
		"interval_ms":            uint64(60000),
		"bar_label":              "open_at",
		"open_at_ms":             openAtMS,
		"finality":               "successor_observed",
		"closed":                 true,
		"regular_session":        true,
		"currency":               "KRW",
		"price_scale":            uint64(0),
		"open_minor":             uint64(71200),
		"high_minor":             uint64(71400),
		"low_minor":              uint64(71100),
		"close_minor":            uint64(71300),
		"volume":                 uint64(5000),
		"raw":                    map[string]any{"open": "71200", "high": "71400", "low": "71100", "close": "71300", "volume": "5000"},
		"revision":               uint64(1),
		"successor_open_at_ms":   openAtMS + 60000,
		"source_observed_at_ms":  openAtMS + 120000,
		"source_response_digest": testResponseDigest,
	}
}

func usTestQuoteFields() map[string]any {
	observed := uint64(time.Date(2026, 8, 14, 14, 0, 0, 0, time.UTC).UnixMilli())
	return map[string]any{
		"schema":                 "official_quote_l1:v1",
		"market":                 "US",
		"symbol":                 "AAPL",
		"currency":               "USD",
		"price_scale":            uint64(4),
		"bid_minor":              uint64(2316000),
		"ask_minor":              uint64(2317000),
		"last_minor":             uint64(2316500),
		"raw":                    map[string]any{"bid": "231.6000", "ask": "231.7000", "last": "231.65"},
		"source_observed_at_ms":  observed,
		"received_at_ms":         observed + 250,
		"source_response_digest": testResponseDigest,
		"revision":               uint64(1),
	}
}

func barPayloadBytes(t *testing.T, fields map[string]any, mutate func(map[string]any)) []byte {
	t.Helper()
	if mutate != nil {
		mutate(fields)
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshalling fixture: %v", err)
	}
	return encoded
}

func quotePayloadBytes(t *testing.T, fields map[string]any, mutate func(map[string]any)) []byte {
	t.Helper()
	return barPayloadBytes(t, fields, mutate)
}

func usTestBarHeader(t *testing.T, revision uint64) Header {
	t.Helper()
	header, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion,
		usTestBarOpenAt, revision, usTestBarOpenAt.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatalf("newClosedBar1mHeader: %v", err)
	}
	return header
}

func usTestQuoteHeader(t *testing.T, revision uint64) Header {
	t.Helper()
	observed := time.Date(2026, 8, 14, 14, 0, 0, 0, time.UTC)
	header, err := newQuoteL1Header(marketclock.MarketUS, "AAPL", observed, revision, observed.Add(250*time.Millisecond), "USD")
	if err != nil {
		t.Fatalf("newQuoteL1Header: %v", err)
	}
	return header
}

func assertBarPayloadRefused(t *testing.T, name string, payload []byte, want string) {
	t.Helper()
	_, err := NewEnvelope(usTestBarHeader(t, 1), payload)
	var refusal *ValidationError
	if !errors.As(err, &refusal) || refusal.Code != RefusalPayloadInvalid {
		t.Fatalf("%s: want payload refusal, got %v", name, err)
	}
	if want != "" && !strings.Contains(refusal.Detail, want) {
		t.Fatalf("%s: refusal detail = %q, want it to mention %q", name, refusal.Detail, want)
	}
}

// ---- P1/P2 후속 규칙 (RED 먼저) ----

// 봉이 자기 세션의 날짜(시장 현지 기준)에 속하는지 본다. 결정 6의 "달력 하루" 묶기다.
func TestClosedBarRequiresTheSessionCalendarDay(t *testing.T) {
	t.Parallel()
	sessionDay := usTestBarOpenAt
	for _, tt := range []struct {
		name   string
		openAt time.Time
	}{
		{"a later month", time.Date(2026, 9, 23, 13, 30, 0, 0, time.UTC)},
		{"a decade earlier", time.Date(1996, 1, 2, 14, 30, 0, 0, time.UTC)},
		{"the next trading day", sessionDay.AddDate(0, 0, 1)},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// 본문 경로: 세션은 US:2026-08-14 인데 봉은 다른 날이다.
			fields := usTestBarFields(uint64(tt.openAt.UnixMilli()))
			header, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion,
				tt.openAt, 1, tt.openAt.Add(2*time.Minute), "USD")
			if err == nil {
				if _, err = NewEnvelope(header, barPayloadBytes(t, fields, nil)); err == nil {
					t.Fatal("a bar from another calendar day was accepted")
				}
			}
			var refusal *ValidationError
			if !errors.As(err, &refusal) {
				t.Fatalf("want a typed refusal, got %v", err)
			}
			// 생성자 경로도 같이 막혀야 한다.
			input := usTestBarInput()
			input.OpenAt = tt.openAt
			input.ObservedAt = tt.openAt.Add(2 * time.Minute)
			input.SuccessorOpenAt = tt.openAt.Add(time.Minute)
			if _, err := NewClosedBar1mEnvelope(input); err == nil {
				t.Fatal("constructor accepted a bar from another calendar day")
			}
		})
	}
}

// 머리말 도우미 자체가 달력 하루를 확인해야 한다. 본문 해독기가 대신 잡아 주는 것에 기대지 않는다.
func TestClosedBar1mHeaderRefusesABarFromAnotherCalendarDay(t *testing.T) {
	t.Parallel()
	for _, openAt := range []time.Time{
		time.Date(2026, 9, 23, 13, 30, 0, 0, time.UTC),
		time.Date(1996, 1, 2, 14, 30, 0, 0, time.UTC),
		usTestBarOpenAt.AddDate(0, 0, 1),
		usTestBarOpenAt.AddDate(0, 0, -1),
	} {
		openAt := openAt
		header, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion,
			openAt, 1, openAt.Add(2*time.Minute), "USD")
		var refusal *ValidationError
		if !errors.As(err, &refusal) || refusal.Code != RefusalIdentityMismatch {
			t.Fatalf("header helper accepted a %s bar inside session %s: %+v (err=%v)", openAt, usTestSessionID, header, err)
		}
	}
}

// 본문 해독기 단독으로도 달력 하루를 확인해야 한다.
// 머리말은 08-14로 멀쩡하고 본문만 08-13 봉을 담은, 머리말 검사가 잡을 수 없는 경우다.
func TestClosedBarPayloadAloneRefusesABarFromAnotherCalendarDay(t *testing.T) {
	t.Parallel()
	for _, offset := range []int{-1, -2, 1} {
		day := usTestBarOpenAt.AddDate(0, 0, offset)
		assertBarPayloadRefused(t, "bar on another day", barPayloadBytes(t, usTestBarFields(uint64(day.UnixMilli())), nil), "session_id")
	}
}

func TestClosedBarAcceptsEveryBarOfTheMarketLocalSessionDay(t *testing.T) {
	t.Parallel()
	// 미국 20:00 ET 시간외 봉은 UTC로 다음 날이지만 뉴욕 현지로는 같은 날이다.
	postMarket := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	input := usTestBarInput()
	input.OpenAt = postMarket
	input.ObservedAt = postMarket.Add(2 * time.Minute)
	input.SuccessorOpenAt = postMarket.Add(time.Minute)
	input.RegularSession = false
	if _, err := NewClosedBar1mEnvelope(input); err != nil {
		t.Fatalf("post-market bar of the same market-local day was refused: %v", err)
	}
	// 한국 09:00 KST = 00:00 UTC 도 같은 세션 날짜다.
	krHeader, err := newClosedBar1mHeader(marketclock.MarketKR, "005930", krTestSessionID, krTestCalendarVersion,
		krTestBarOpenAt, 1, krTestBarOpenAt.Add(2*time.Minute), "KRW")
	if err != nil {
		t.Fatalf("KR 09:00 KST bar refused: %v", err)
	}
	if _, err := NewEnvelope(krHeader, barPayloadBytes(t, krTestBarFields(uint64(krTestBarOpenAt.UnixMilli())), nil)); err != nil {
		t.Fatalf("KR 09:00 KST payload refused: %v", err)
	}
}

// 닫힘의 최소 조건: 관측 시각이 적어도 봉이 끝난 시각 이후여야 한다.
func TestClosedBarRequiresObservationAfterTheBarClosed(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	assertBarPayloadRefused(t, "one millisecond early", barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["source_observed_at_ms"] = base + 59999
	}), "source_observed_at_ms")
	accepted, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["source_observed_at_ms"] = base + 60000
	}))
	if err != nil {
		t.Fatalf("observation exactly at the bar close was refused: %v", err)
	}
	if decoded, err := DecodeClosedBar1mPayload(accepted.CanonicalPayload()); err != nil || decoded.SourceObservedAtMS != base+60000 {
		t.Fatalf("decoded = %+v, err = %v", decoded, err)
	}
}

func TestClosedBarRefusesSymbolWithRecordSeparatorOrLowerCase(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	for _, symbol := range []string{"AA:PL", ":AAPL", "AAPL:", "aapl", "Aapl"} {
		symbol := symbol
		t.Run(symbol, func(t *testing.T) {
			t.Parallel()
			assertBarPayloadRefused(t, "symbol "+symbol, barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
				fields["symbol"] = symbol
			}), "symbol")
		})
	}
	if _, err := newClosedBar1mHeader(marketclock.MarketUS, "AA:PL", usTestSessionID, usTestCalendarVersion,
		usTestBarOpenAt, 1, usTestBarOpenAt.Add(2*time.Minute), "USD"); err == nil {
		t.Fatal("header helper accepted a symbol carrying the record-id separator")
	}
}

func TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		raw   string
		scale uint64
		want  uint64
	}{
		{"0", 0, 0},
		{"0", 4, 0},
		{"0.5", 4, 5000},
		{"0.0001", 4, 1},
		{"1844674407370955.1615", 4, 18446744073709551615},
		{"231.1", 4, 2311000},
	} {
		tt := tt
		got, err := minorFromRawDecimal(tt.raw, tt.scale)
		if err != nil || got != tt.want {
			t.Fatalf("minorFromRawDecimal(%q, %d) = %d, %v; want %d", tt.raw, tt.scale, got, err, tt.want)
		}
	}
	for _, tt := range []struct {
		raw   string
		scale uint64
	}{
		{"00231.4350", 4},
		{"01", 0},
		{"00", 0},
		{"1844674407370955.1616", 4},
		{"18446744073709551616", 4},
		{"18446744073709551616", 0},
		{strings.Repeat("1", 33), 0},
	} {
		tt := tt
		if got, err := minorFromRawDecimal(tt.raw, tt.scale); err == nil {
			t.Fatalf("minorFromRawDecimal(%q, %d) = %d, want a refusal", tt.raw, tt.scale, got)
		}
	}
	// 길이 상한(32자)을 없앤 뒤에도 긴 값은 막힌다. 다만 막는 규칙이 무엇인지 분명히 해 둔다.
	// 33자짜리 이 값은 소수 4자리라 scale 규칙은 통과하고, 정수부 28자리에서 자릿수 넘침으로 걸린다.
	if _, err := minorFromRawDecimal("1234567890123456789012345678.1234", 4); err == nil {
		t.Fatal("a 33 character raw was accepted")
	} else if !strings.Contains(err.Error(), "does not fit in 64 bits") {
		t.Fatalf("33 character raw refused by %q, want the 64 bit overflow rule", err)
	}
	// 소수 자릿수가 넘치는 긴 값은 scale 규칙이 먼저 막는다.
	if _, err := minorFromRawDecimal("231.123456789012345678901234567890", 4); err == nil {
		t.Fatal("an over-precise long raw was accepted")
	} else if !strings.Contains(err.Error(), "fraction digits") {
		t.Fatalf("over-precise long raw refused by %q, want the scale rule", err)
	}
}

func TestClosedBarRefusesLeadingZeroRawDecimal(t *testing.T) {
	t.Parallel()
	assertBarPayloadRefused(t, "leading zero close", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["close"] = "0231.65"
	}), "close")
	assertBarPayloadRefused(t, "leading zero volume", barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["raw"].(map[string]any)["volume"] = "012345"
	}), "volume")
}

func TestCanonicalIntegerValueBoundaries(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"0", "1", "1e19", "18446744073709551615", "6e4"} {
		if _, err := canonicalIntegerValue(text); err != nil {
			t.Fatalf("canonicalIntegerValue(%q) refused: %v", text, err)
		}
	}
	for _, text := range []string{"2e19", "18446744073709551616", "-1", "15e-1", "1e21", ""} {
		if got, err := canonicalIntegerValue(text); err == nil {
			t.Fatalf("canonicalIntegerValue(%q) = %d, want a refusal", text, got)
		}
	}
	base := uint64(usTestBarOpenAt.UnixMilli())
	accepted, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["volume"] = json.RawMessage("1e19")
		fields["raw"].(map[string]any)["volume"] = "10000000000000000000"
	}))
	if err != nil {
		t.Fatalf("volume 1e19 refused: %v", err)
	}
	if decoded, _ := DecodeClosedBar1mPayload(accepted.CanonicalPayload()); decoded.Volume != 10000000000000000000 {
		t.Fatalf("volume = %d", decoded.Volume)
	}
	assertBarPayloadRefused(t, "volume above 2^64-1", barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["volume"] = json.RawMessage("2e19")
		fields["raw"].(map[string]any)["volume"] = "20000000000000000000"
	}), "volume")
}

func TestDecodePayloadRefusesNonCanonicalJSON(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	canonical := barPayloadBytes(t, usTestBarFields(base), nil)
	duplicated := append([]byte(nil), canonical[:len(canonical)-1]...)
	// 값을 바꾸지 않고 같은 열쇠만 한 번 더 붙인다. 그래야 "중복 열쇠" 자체가 시험된다.
	duplicated = append(duplicated, []byte(`,"volume":12345}`)...)
	if _, err := DecodeClosedBar1mPayload(duplicated); err == nil {
		t.Fatal("DecodeClosedBar1mPayload accepted a duplicate JSON key")
	}
	trailing := append(append([]byte(nil), canonical...), []byte(`{"another":1}`)...)
	if _, err := DecodeClosedBar1mPayload(trailing); err == nil {
		t.Fatal("DecodeClosedBar1mPayload accepted a trailing JSON value")
	}
	quote := quotePayloadBytes(t, usTestQuoteFields(), nil)
	quoteDuplicated := append([]byte(nil), quote[:len(quote)-1]...)
	quoteDuplicated = append(quoteDuplicated, []byte(`,"revision":1}`)...)
	if _, err := DecodeQuoteL1Payload(quoteDuplicated); err == nil {
		t.Fatal("DecodeQuoteL1Payload accepted a duplicate JSON key")
	}
	quoteTrailing := append(append([]byte(nil), quote...), []byte(`[]`)...)
	if _, err := DecodeQuoteL1Payload(quoteTrailing); err == nil {
		t.Fatal("DecodeQuoteL1Payload accepted a trailing JSON value")
	}
}

// 결정 6의 successor-observed 주장을 본문이 직접 들고 다닌다.
// L1a는 "구조적으로 말이 되는가"만 본다. 진짜로 뒤 봉을 봤는지는 L1b 생산자의 의무다.
func TestClosedBarRequiresSuccessorOpenAt(t *testing.T) {
	t.Parallel()
	base := uint64(usTestBarOpenAt.UnixMilli())
	accepted, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(base), nil))
	if err != nil {
		t.Fatalf("bar carrying a successor was refused: %v", err)
	}
	decoded, err := DecodeClosedBar1mPayload(accepted.CanonicalPayload())
	if err != nil {
		t.Fatalf("replay refused: %v", err)
	}
	if decoded.SuccessorOpenAtMS != base+60000 {
		t.Fatalf("successor_open_at_ms = %d, want %d", decoded.SuccessorOpenAtMS, base+60000)
	}
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"not on the minute", base + 60001},
		{"at the bar itself", base},
		{"before the bar", base - 60000},
		{"after the observation instant", base + 180000},
		{"zero", uint64(0)},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertBarPayloadRefused(t, tt.name, barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
				fields["successor_open_at_ms"] = tt.value
			}), "successor_open_at_ms")
		})
	}
	assertBarPayloadRefused(t, "missing", barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		delete(fields, "successor_open_at_ms")
	}), "successor_open_at_ms")
	// 관측 시각과 정확히 같은 값은 허용된다 (뒤 봉이 열린 순간에 본 경우).
	if _, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(base), func(fields map[string]any) {
		fields["successor_open_at_ms"] = base + 120000
	})); err != nil {
		t.Fatalf("successor exactly at the observation instant was refused: %v", err)
	}
}

func TestNewClosedBar1mEnvelopeCarriesTheObservedSuccessor(t *testing.T) {
	t.Parallel()
	input := usTestBarInput()
	envelope, err := NewClosedBar1mEnvelope(input)
	if err != nil {
		t.Fatalf("constructor refused: %v", err)
	}
	bar, err := DecodeClosedBar1mPayload(envelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("replay refused: %v", err)
	}
	if bar.SuccessorOpenAtMS != uint64(input.SuccessorOpenAt.UnixMilli()) {
		t.Fatalf("successor = %d, want %d", bar.SuccessorOpenAtMS, input.SuccessorOpenAt.UnixMilli())
	}
	for _, tt := range []struct {
		name      string
		successor time.Time
	}{
		{"zero", time.Time{}},
		{"the bar itself", input.OpenAt},
		{"not on the minute", input.OpenAt.Add(90 * time.Second)},
		{"after the observation instant", input.ObservedAt.Add(time.Minute)},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mutated := usTestBarInput()
			mutated.SuccessorOpenAt = tt.successor
			if _, err := NewClosedBar1mEnvelope(mutated); err == nil {
				t.Fatalf("%s: constructor accepted an impossible successor", tt.name)
			}
		})
	}
}
