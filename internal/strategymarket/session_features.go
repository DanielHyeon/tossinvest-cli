package strategymarket

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
	"time"
)

type FeatureState string

const (
	FeatureKnown       FeatureState = "known"
	FeatureNotReady    FeatureState = "not_ready"
	FeatureInvalid     FeatureState = "invalid"
	FeatureUnsupported FeatureState = "unsupported"
)

type SessionFeatureBar struct {
	Market, Symbol, Currency, Source, Basis string
	CloseAt, ReceivedAt                     time.Time
	Open, High, Low, Close, Volume          string
}
type SessionFeatureInput struct {
	Market, Symbol, Currency, TradingDate, CalendarSource, CalendarVersion string
	CalendarVerified                                                       bool
	OpenAt, CloseAt, AsOf                                                  time.Time
	Minutes, Fives                                                         []SessionFeatureBar
}
type FeatureValue struct {
	State         FeatureState
	Value, Reason string
}
type SessionFeatures struct {
	Status                                                                                                                                     FeatureState
	Reason, Method, EMAMethod, SlopeMethod, Market, Symbol, Currency, TradingDate, CalendarSource, CalendarVersion, Source, Provenance, Digest string
	AsOf, OpenAt, CloseAt, FirstMinuteAt, LastMinuteAt, FirstFiveAt, LastFiveAt, LastReceivedAt, SlopePreviousAt, SlopeCurrentAt               time.Time
	MinuteCount, FiveCount                                                                                                                     int
	VWAP, EMA9, Slope                                                                                                                          FeatureValue
}

// CalculateSessionFeatures는 호출자가 제공한 미인증 세션·봉을 순수 계산함.
func CalculateSessionFeatures(in SessionFeatureInput) SessionFeatures {
	out := SessionFeatures{Status: FeatureNotReady, Method: "session_typical_price_1m_v1", EMAMethod: "session_close_5m_sma9_alpha_1_5_v1", SlopeMethod: "consecutive_5m_vwap_percent_v1", Market: in.Market, Symbol: in.Symbol, Currency: in.Currency, TradingDate: in.TradingDate, CalendarSource: in.CalendarSource, CalendarVersion: in.CalendarVersion, OpenAt: in.OpenAt, CloseAt: in.CloseAt, Source: string(SourceOfficialOpenAPI), Provenance: "caller-supplied, unverified", AsOf: in.AsOf, MinuteCount: len(in.Minutes), FiveCount: len(in.Fives)}
	no := FeatureValue{State: FeatureNotReady, Reason: "incomplete evidence"}
	out.VWAP = no
	out.EMA9 = no
	out.Slope = no
	fail := func(state FeatureState, reason string) SessionFeatures {
		out.Status = state
		out.Reason = reason
		return out
	}
	if in.Market == "" {
		return fail(FeatureNotReady, "missing market")
	}
	if in.Market != "KR" {
		return fail(FeatureUnsupported, "market calendar unsupported")
	}
	if !in.CalendarVerified || in.Symbol == "" || in.Currency == "" || in.TradingDate == "" || in.CalendarVersion == "" || in.CalendarSource == "" || in.OpenAt.IsZero() || in.CloseAt.IsZero() || in.AsOf.IsZero() {
		return fail(FeatureNotReady, "missing calendar or session evidence")
	}
	zone := time.FixedZone("KST", 9*60*60)
	open := in.OpenAt.In(zone)
	closeTime := in.CloseAt.In(zone)
	duration := in.CloseAt.Sub(in.OpenAt)
	if in.Currency != "KRW" || in.CalendarSource != string(SourceOfficialOpenAPI) || strings.TrimSpace(in.Symbol) != in.Symbol || strings.TrimSpace(in.CalendarVersion) != in.CalendarVersion || open.Format("2006-01-02") != in.TradingDate || closeTime.Format("2006-01-02") != in.TradingDate || open.Second() != 0 || open.Nanosecond() != 0 || duration <= 0 || duration > 390*time.Minute || duration%(5*time.Minute) != 0 || len(in.Minutes) > 390 || len(in.Fives) > 78 {
		return fail(FeatureInvalid, "invalid calendar or session bounds")
	}
	if in.AsOf.Before(in.OpenAt) {
		if len(in.Minutes) != 0 || len(in.Fives) != 0 {
			return fail(FeatureInvalid, "bars before session opened")
		}
		return fail(FeatureNotReady, "session has not opened")
	}
	// 정수 초과 및 비정상 크기 입력을 제한하고 연산은 유리수로 수행함.
	type parsed struct {
		values [5]*big.Rat
		at     time.Time
	}
	parse := func(bars []SessionFeatureBar, interval time.Duration) ([]parsed, FeatureState, string) {
		parsedBars := make([]parsed, 0, len(bars))
		var prev time.Time
		for _, b := range bars {
			if b.Market == "" || b.Symbol == "" || b.Currency == "" || b.Source == "" || b.Basis == "" || b.CloseAt.IsZero() || b.ReceivedAt.IsZero() {
				return nil, FeatureNotReady, "missing bar evidence"
			}
			elapsed := b.CloseAt.Sub(in.OpenAt)
			if b.Market != in.Market || b.Symbol != in.Symbol || b.Currency != in.Currency || b.Source != string(SourceOfficialOpenAPI) || b.Basis != "unadjusted" || elapsed <= 0 || elapsed > duration || elapsed%interval != 0 || b.CloseAt.After(in.AsOf) || b.ReceivedAt.Before(b.CloseAt) || b.ReceivedAt.After(in.AsOf) || (!prev.IsZero() && !b.CloseAt.After(prev)) {
				return nil, FeatureInvalid, "invalid bar identity, order or timing"
			}
			p := parsed{at: b.CloseAt}
			fields := []string{b.Open, b.High, b.Low, b.Close, b.Volume}
			for j, s := range fields {
				if !boundedFeatureDecimal(s) {
					return nil, FeatureInvalid, "numeric input too large"
				}
				v, ok := exactDecimal(s)
				if !ok || (j < 4 && v.Sign() <= 0) || (j == 4 && v.Sign() < 0) {
					return nil, FeatureInvalid, "invalid OHLCV"
				}
				p.values[j] = v
			}
			if p.values[1].Cmp(p.values[0]) < 0 || p.values[1].Cmp(p.values[2]) < 0 || p.values[1].Cmp(p.values[3]) < 0 || p.values[2].Cmp(p.values[0]) > 0 || p.values[2].Cmp(p.values[3]) > 0 {
				return nil, FeatureInvalid, "invalid OHLC range"
			}
			parsedBars = append(parsedBars, p)
			prev = b.CloseAt
			if b.ReceivedAt.After(out.LastReceivedAt) {
				out.LastReceivedAt = b.ReceivedAt
			}
		}
		return parsedBars, FeatureKnown, ""
	}
	minutes, state, reason := parse(in.Minutes, time.Minute)
	if state != FeatureKnown {
		return fail(state, reason)
	}
	fives, state, reason := parse(in.Fives, 5*time.Minute)
	if state != FeatureKnown {
		return fail(state, reason)
	}
	if len(fives) > 0 {
		out.FirstFiveAt = fives[0].at
		out.LastFiveAt = fives[len(fives)-1].at
	}
	if len(minutes) > 0 {
		out.FirstMinuteAt = minutes[0].at
		out.LastMinuteAt = minutes[len(minutes)-1].at
	}
	// 모든 제공된 5분봉을 같은 시각의 1분봉과 대조함.
	byTime := make(map[time.Time]parsed, len(minutes))
	for _, m := range minutes {
		byTime[m.at.UTC()] = m
	}
	for _, five := range fives {
		bucket := make([]parsed, 0, 5)
		for j := 4; j >= 0; j-- {
			m, ok := byTime[five.at.Add(-time.Duration(j)*time.Minute).UTC()]
			if !ok {
				bucket = nil
				break
			}
			bucket = append(bucket, m)
		}
		if len(bucket) == 0 {
			continue
		}
		high := new(big.Rat).Set(bucket[0].values[1])
		low := new(big.Rat).Set(bucket[0].values[2])
		volume := new(big.Rat)
		for _, m := range bucket {
			if m.values[1].Cmp(high) > 0 {
				high.Set(m.values[1])
			}
			if m.values[2].Cmp(low) < 0 {
				low.Set(m.values[2])
			}
			volume.Add(volume, m.values[4])
		}
		expected := []*big.Rat{bucket[0].values[0], high, low, bucket[4].values[3], volume}
		for j, v := range expected {
			if v.Cmp(five.values[j]) != 0 {
				return fail(FeatureInvalid, "five-minute OHLCV mismatch")
			}
		}
	}
	// Digest는 동일한 값과 시각의 표현 차이를 제거해 재생을 고정함.
	canonical := in
	canonical.OpenAt = in.OpenAt.UTC()
	canonical.CloseAt = in.CloseAt.UTC()
	canonical.AsOf = in.AsOf.UTC()
	canonical.Minutes = append([]SessionFeatureBar(nil), in.Minutes...)
	canonical.Fives = append([]SessionFeatureBar(nil), in.Fives...)
	normalize := func(bars []SessionFeatureBar, parsedBars []parsed) {
		for i := range bars {
			b := &bars[i]
			b.CloseAt = b.CloseAt.UTC()
			b.ReceivedAt = b.ReceivedAt.UTC()
			v := parsedBars[i].values
			b.Open = v[0].RatString()
			b.High = v[1].RatString()
			b.Low = v[2].RatString()
			b.Close = v[3].RatString()
			b.Volume = v[4].RatString()
		}
	}
	normalize(canonical.Minutes, minutes)
	normalize(canonical.Fives, fives)
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return fail(FeatureInvalid, "unrepresentable canonical input")
	}
	digest := sha256.Sum256(encoded)
	out.Digest = hex.EncodeToString(digest[:])
	elapsed := in.AsOf.Sub(in.OpenAt)
	if elapsed > duration {
		elapsed = duration
	}
	need := int(elapsed / time.Minute)
	if len(minutes) != need {
		out.Reason = "incomplete minute prefix"
		return out
	}
	for i, m := range minutes {
		if !m.at.Equal(in.OpenAt.Add(time.Duration(i+1) * time.Minute)) {
			out.Reason = "incomplete minute prefix"
			return out
		}
	}
	sum, volume := new(big.Rat), new(big.Rat)
	var prevVWAP, currentVWAP *big.Rat
	for i, m := range minutes {
		typical := new(big.Rat).Quo(new(big.Rat).Add(new(big.Rat).Add(m.values[1], m.values[2]), m.values[3]), big.NewRat(3, 1))
		sum.Add(sum, new(big.Rat).Mul(typical, m.values[4]))
		volume.Add(volume, m.values[4])
		if (i+1)%5 == 0 {
			prevVWAP = currentVWAP
			currentVWAP = nil
			if volume.Sign() > 0 {
				currentVWAP = new(big.Rat).Quo(sum, volume)
			}
		}
	}
	if volume.Sign() > 0 {
		out.VWAP = FeatureValue{State: FeatureKnown, Value: new(big.Rat).Quo(sum, volume).RatString()}
	} else {
		out.VWAP.Reason = "zero cumulative volume"
	}
	needFives := need / 5
	if len(fives) != needFives {
		out.Reason = "incomplete five-minute prefix"
		return out
	}
	for i, f := range fives {
		if !f.at.Equal(in.OpenAt.Add(time.Duration(i+1) * 5 * time.Minute)) {
			out.Reason = "incomplete five-minute prefix"
			return out
		}
	}
	if needFives >= 9 {
		ema := new(big.Rat)
		for _, f := range fives[:9] {
			ema.Add(ema, f.values[3])
		}
		ema.Quo(ema, big.NewRat(9, 1))
		for _, f := range fives[9:] {
			ema.Add(new(big.Rat).Mul(ema, big.NewRat(4, 5)), new(big.Rat).Mul(f.values[3], big.NewRat(1, 5)))
		}
		out.EMA9 = FeatureValue{State: FeatureKnown, Value: ema.RatString()}
	} else {
		out.EMA9.Reason = "fewer than nine five-minute closes"
	}
	if needFives >= 2 && prevVWAP != nil && currentVWAP != nil && prevVWAP.Sign() > 0 {
		slope := new(big.Rat).Mul(new(big.Rat).Sub(new(big.Rat).Quo(currentVWAP, prevVWAP), big.NewRat(1, 1)), big.NewRat(100, 1))
		out.Slope = FeatureValue{State: FeatureKnown, Value: slope.RatString()}
		out.SlopeCurrentAt = in.OpenAt.Add(time.Duration(needFives) * 5 * time.Minute)
		out.SlopePreviousAt = out.SlopeCurrentAt.Add(-5 * time.Minute)
	} else {
		out.Slope.Reason = "two positive boundary VWAPs unavailable"
	}
	if out.VWAP.State == FeatureKnown && out.EMA9.State == FeatureKnown && out.Slope.State == FeatureKnown {
		out.Status = FeatureKnown
	}
	return out
}

func boundedFeatureDecimal(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	digits, dot := 0, false
	for i, c := range s {
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c == '.' && !dot:
			dot = true
		case (c == '+' || c == '-') && i == 0:
		default:
			return false
		}
	}
	return digits > 0
}
