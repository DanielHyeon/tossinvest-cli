package strategymarket

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func featureFixture(n int) SessionFeatureInput {
	z := time.FixedZone("KST", 9*3600)
	open := time.Date(2026, 5, 8, 9, 0, 0, 0, z)
	in := SessionFeatureInput{Market: "KR", Symbol: "005930", Currency: "KRW", CalendarVersion: "v1", CalendarSource: "official-open-api", CalendarVerified: true, TradingDate: "2026-05-08", OpenAt: open, CloseAt: open.Add(390 * time.Minute), AsOf: open.Add(time.Duration(n) * time.Minute)}
	for i := 1; i <= n; i++ {
		p := strconv.Itoa(101 + (i-1)/5)
		at := open.Add(time.Duration(i) * time.Minute)
		in.Minutes = append(in.Minutes, SessionFeatureBar{Market: in.Market, Symbol: in.Symbol, Currency: in.Currency, Source: "official-open-api", Basis: "unadjusted", CloseAt: at, ReceivedAt: at, Open: p, High: p, Low: p, Close: p, Volume: "1"})
		if i%5 == 0 {
			in.Fives = append(in.Fives, SessionFeatureBar{Market: in.Market, Symbol: in.Symbol, Currency: in.Currency, Source: "official-open-api", Basis: "unadjusted", CloseAt: at, ReceivedAt: at, Open: p, High: p, Low: p, Close: p, Volume: "5"})
		}
	}
	return in
}
func TestSessionFeaturesGoldenAndReplay(t *testing.T) {
	in := featureFixture(50)
	before := featureFixture(50)
	got := CalculateSessionFeatures(in)
	if got.Status != FeatureKnown || got.VWAP.Value != "211/2" || got.EMA9.Value != "106" || got.Slope.Value != "10/21" {
		t.Fatalf("%+v", got)
	}
	if got.Method != "session_typical_price_1m_v1" || got.Digest == "" || got.MinuteCount != 50 || got.FiveCount != 10 || !got.SlopeCurrentAt.Equal(in.OpenAt.Add(50*time.Minute)) || !got.SlopePreviousAt.Equal(in.OpenAt.Add(45*time.Minute)) {
		t.Fatalf("metadata: %+v", got)
	}
	if !reflect.DeepEqual(in, before) || !reflect.DeepEqual(got, CalculateSessionFeatures(in)) {
		t.Fatal("input changed or replay differs")
	}
	in = featureFixture(45)
	got = CalculateSessionFeatures(in)
	if got.EMA9.Value != "105" || got.EMA9.State != FeatureKnown {
		t.Fatalf("seed: %+v", got)
	}
	in = featureFixture(46)
	got = CalculateSessionFeatures(in)
	if got.Slope.Value != "100/209" || !got.SlopeCurrentAt.Equal(in.OpenAt.Add(45*time.Minute)) || !got.SlopePreviousAt.Equal(in.OpenAt.Add(40*time.Minute)) {
		t.Fatalf("boundaries: %+v", got)
	}
	in = featureFixture(1)
	got = CalculateSessionFeatures(in)
	if got.VWAP.State != FeatureKnown || got.EMA9.State != FeatureNotReady || got.Slope.State != FeatureNotReady {
		t.Fatalf("reset: %+v", got)
	}
}
func TestSessionFeaturesBoundariesAndReadiness(t *testing.T) {
	for _, n := range []int{0, 1, 5, 44, 45, 390} {
		in := featureFixture(n)
		got := CalculateSessionFeatures(in)
		want := FeatureNotReady
		if n >= 45 {
			want = FeatureKnown
		}
		if got.Status != want {
			t.Fatalf("%d: %+v", n, got)
		}
		if n < 45 && got.EMA9.State != FeatureNotReady {
			t.Fatalf("%d: %+v", n, got)
		}
	}
	in := featureFixture(45)
	in.Minutes = in.Minutes[25:]
	got := CalculateSessionFeatures(in)
	if got.VWAP.State != FeatureNotReady || got.EMA9.State != FeatureNotReady {
		t.Fatalf("partial prefix %+v", got)
	}
	in = featureFixture(5)
	for i := range in.Minutes {
		in.Minutes[i].Volume = "0"
	}
	in.Fives[0].Volume = "0"
	got = CalculateSessionFeatures(in)
	if got.VWAP.State != FeatureNotReady {
		t.Fatalf("zero volume %+v", got)
	}
	in = featureFixture(45)
	for i := range in.Minutes {
		in.Minutes[i].Volume = "0"
	}
	for i := range in.Fives {
		in.Fives[i].Volume = "0"
	}
	got = CalculateSessionFeatures(in)
	if got.Status != FeatureNotReady || got.VWAP.State != FeatureNotReady || got.EMA9.State != FeatureKnown || got.Slope.State != FeatureNotReady {
		t.Fatalf("zero volume with EMA seed: %+v", got)
	}
}
func TestSessionFeaturesRefusesInvalidEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SessionFeatureInput)
		want   FeatureState
	}{
		{"US unsupported", func(i *SessionFeatureInput) { i.Market = "US" }, FeatureUnsupported},
		{"missing market", func(i *SessionFeatureInput) { i.Market = "" }, FeatureNotReady},
		{"bars before session", func(i *SessionFeatureInput) { i.AsOf = i.OpenAt.Add(-time.Minute) }, FeatureInvalid},
		{"unverified calendar", func(i *SessionFeatureInput) { i.CalendarVerified = false }, FeatureNotReady},
		{"missing receive", func(i *SessionFeatureInput) { i.Minutes[0].ReceivedAt = time.Time{} }, FeatureNotReady},
		{"received before close", func(i *SessionFeatureInput) { i.Minutes[0].ReceivedAt = i.Minutes[0].CloseAt.Add(-time.Second) }, FeatureInvalid},
		{"received future", func(i *SessionFeatureInput) { i.Minutes[0].ReceivedAt = i.AsOf.Add(time.Second) }, FeatureInvalid},
		{"future bar", func(i *SessionFeatureInput) { i.Minutes[0].CloseAt = i.AsOf.Add(time.Minute) }, FeatureInvalid},
		{"duplicate", func(i *SessionFeatureInput) { i.Minutes[1] = i.Minutes[0] }, FeatureInvalid},
		{"reverse", func(i *SessionFeatureInput) { i.Minutes[0], i.Minutes[1] = i.Minutes[1], i.Minutes[0] }, FeatureInvalid},
		{"open label", func(i *SessionFeatureInput) { i.Minutes[0].CloseAt = i.OpenAt }, FeatureInvalid},
		{"mixed symbol", func(i *SessionFeatureInput) { i.Minutes[0].Symbol = "other" }, FeatureInvalid},
		{"mixed currency", func(i *SessionFeatureInput) { i.Fives[0].Currency = "USD" }, FeatureInvalid},
		{"adjusted", func(i *SessionFeatureInput) { i.Minutes[0].Basis = "adjusted" }, FeatureInvalid},
		{"negative volume", func(i *SessionFeatureInput) { i.Minutes[0].Volume = "-1" }, FeatureInvalid},
		{"binary exponent", func(i *SessionFeatureInput) { i.Minutes[0].Open = "1p10000000" }, FeatureInvalid},
		{"hexadecimal", func(i *SessionFeatureInput) { i.Minutes[0].Open = "0x10" }, FeatureInvalid},
		{"underscore", func(i *SessionFeatureInput) { i.Minutes[0].Open = "1_00" }, FeatureInvalid},
		{"invalid OHLC", func(i *SessionFeatureInput) { i.Minutes[0].High = "1" }, FeatureInvalid},
		{"mismatched five", func(i *SessionFeatureInput) { i.Fives[0].Close = "999"; i.Fives[0].High = "999" }, FeatureInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := featureFixture(45)
			tc.mutate(&in)
			got := CalculateSessionFeatures(in)
			if got.Status != tc.want || got.VWAP.Value != "" || got.EMA9.Value != "" || got.Slope.Value != "" {
				t.Fatalf("%+v", got)
			}
		})
	}
}

func TestSessionFeaturesWeightedTypicalPriceAndNewSession(t *testing.T) {
	in := featureFixture(10)
	for i := range in.Minutes {
		if i < 5 {
			in.Minutes[i].Open, in.Minutes[i].High, in.Minutes[i].Low, in.Minutes[i].Close = "100", "103", "99", "101"
		} else {
			in.Minutes[i].Open, in.Minutes[i].High, in.Minutes[i].Low, in.Minutes[i].Close, in.Minutes[i].Volume = "102", "109", "101", "107", "2"
		}
	}
	in.Fives[0].Open, in.Fives[0].High, in.Fives[0].Low, in.Fives[0].Close = "100", "103", "99", "101"
	in.Fives[1].Open, in.Fives[1].High, in.Fives[1].Low, in.Fives[1].Close, in.Fives[1].Volume = "102", "109", "101", "107", "10"
	got := CalculateSessionFeatures(in)
	if got.VWAP.Value != "937/9" || got.Slope.Value != "2800/909" || got.Status != FeatureNotReady {
		t.Fatalf("weighted typical price: %+v", got)
	}
	next := featureFixture(1)
	next.TradingDate = "2026-05-11"
	next.OpenAt, next.CloseAt, next.AsOf = next.OpenAt.AddDate(0, 0, 3), next.CloseAt.AddDate(0, 0, 3), next.AsOf.AddDate(0, 0, 3)
	next.Minutes[0].CloseAt, next.Minutes[0].ReceivedAt = next.Minutes[0].CloseAt.AddDate(0, 0, 3), next.Minutes[0].ReceivedAt.AddDate(0, 0, 3)
	got = CalculateSessionFeatures(next)
	if got.VWAP.Value != "101" || got.EMA9.State != FeatureNotReady || got.Status != FeatureNotReady {
		t.Fatalf("new session: %+v", got)
	}
}

func TestSessionFeaturesPrefixAndBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*SessionFeatureInput)
		want   FeatureState
	}{
		{"minute gap", func(i *SessionFeatureInput) { i.Minutes = append(i.Minutes[:8], i.Minutes[9:]...) }, FeatureNotReady},
		{"missing last minute", func(i *SessionFeatureInput) { i.Minutes = i.Minutes[:len(i.Minutes)-1] }, FeatureNotReady},
		{"missing five-minute", func(i *SessionFeatureInput) { i.Fives = i.Fives[:len(i.Fives)-1] }, FeatureNotReady},
		{"five duplicate", func(i *SessionFeatureInput) { i.Fives[1] = i.Fives[0] }, FeatureInvalid},
		{"five reverse", func(i *SessionFeatureInput) { i.Fives[0], i.Fives[1] = i.Fives[1], i.Fives[0] }, FeatureInvalid},
		{"mixed session", func(i *SessionFeatureInput) { i.Minutes[0].CloseAt = i.Minutes[0].CloseAt.Add(24 * time.Hour) }, FeatureInvalid},
		{"late close", func(i *SessionFeatureInput) { i.Minutes[len(i.Minutes)-1].CloseAt = i.CloseAt.Add(time.Minute) }, FeatureInvalid},
		{"missing calendar version", func(i *SessionFeatureInput) { i.CalendarVersion = "" }, FeatureNotReady},
		{"wrong trading date", func(i *SessionFeatureInput) { i.TradingDate = "2026-05-09" }, FeatureInvalid},
		{"unbounded numeric", func(i *SessionFeatureInput) {
			i.Minutes[0].Open = "12345678901234567890123456789012345678901234567890123456789012345"
		}, FeatureInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := featureFixture(45)
			tc.mutate(&in)
			got := CalculateSessionFeatures(in)
			if tc.want == FeatureNotReady {
				if got.EMA9.State != FeatureNotReady || got.EMA9.Value != "" {
					t.Fatalf("%+v", got)
				}
			}
			if got.Status != tc.want {
				t.Fatalf("%+v", got)
			}
		})
	}
}
func TestSessionFeaturesCanonicalDigestAndZeroVolume(t *testing.T) {
	a := featureFixture(10)
	b := featureFixture(10)
	for i := range b.Minutes {
		b.Minutes[i].CloseAt = b.Minutes[i].CloseAt.UTC()
		b.Minutes[i].ReceivedAt = b.Minutes[i].ReceivedAt.UTC()
		b.Minutes[i].Open += ".0"
	}
	got := CalculateSessionFeatures(a)
	other := CalculateSessionFeatures(b)
	if got.Digest == "" || got.Digest != other.Digest {
		t.Fatalf("digest mismatch: %s %s", got.Digest, other.Digest)
	}
	for i := 5; i < 10; i++ {
		a.Minutes[i].Volume = "0"
	}
	a.Fives[1].Volume = "0"
	got = CalculateSessionFeatures(a)
	if got.VWAP.Value != "101" || got.Slope.State != FeatureKnown || got.Slope.Value != "0" {
		t.Fatalf("zero-volume bucket: %+v", got)
	}
}
