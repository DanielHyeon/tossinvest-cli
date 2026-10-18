package console

import "testing"

func TestA154HypotheticalExactGeometry(t *testing.T) {
	in := hypotheticalInput{Market: "KR", Entry: "10000", Low: "9900", ATR: "80", Tick: "10", Min: "1", Max: "20000", Budget: "7000", MaxShares: "10000", MaxNotional: "10000000"}
	got := calculateHypothetical(in)
	wants := [][3]string{{"9930", "100", "7000"}, {"9890", "63", "6930"}, {"9880", "58", "6960"}}
	for i, want := range wants {
		a := got.Arms[i]
		if a.Stop != want[0] || a.Quantity != want[1] || a.Risk != want[2] {
			t.Errorf("arm %d = %+v, want %v", i, a, want)
		}
	}
}

func TestA154IndependentNumericalOracles(t *testing.T) {
	cases := []struct {
		in   hypotheticalInput
		want [][4]string
	}{
		{hypotheticalInput{Market: "KR", Entry: "70000", Low: "69000", ATR: "500", Tick: "100", Min: "60000", Max: "80000", Budget: "10000", MaxShares: "100", MaxNotional: "7000000"}, [][4]string{{"69500", "20", "10000", "1400000"}, {"68900", "9", "9900", "630000"}, {"69200", "12", "9600", "840000"}}},
		{hypotheticalInput{Market: "US", Entry: "100", Low: "98", ATR: "2", Tick: "0.01", Min: "1", Max: "200", Budget: "100", MaxShares: "1000", MaxNotional: "100000"}, [][4]string{{"99.3", "142", "99.4", "14200"}, {"97.99", "49", "98.49", "4900"}, {"97", "33", "99", "3300"}}},
	}
	for _, tc := range cases {
		got := calculateHypothetical(tc.in)
		for i, w := range tc.want {
			a := got.Arms[i]
			if a.Stop != w[0] || a.Quantity != w[1] || a.Risk != w[2] || a.Notional != w[3] {
				t.Errorf("%s arm %d got %+v want %v", tc.in.Market, i, a, w)
			}
		}
	}
}
func TestA154UnknownInvalidAndZeroQuantity(t *testing.T) {
	in := hypotheticalInput{Market: "US", Entry: "100", Low: "101", ATR: "", Tick: "0.01", Min: "1", Max: "200", Budget: "0.01", MaxShares: "10", MaxNotional: "1000"}
	got := calculateHypothetical(in)
	if got.Arms[0].Stop != "99.3" || got.Arms[0].Quantity != "0" || got.Arms[0].State != "매수 가능 수량 없음" || got.Arms[1].State != "입력 오류" || got.Arms[2].Stop != "" || got.Errors["low"] == "" {
		t.Fatalf("unexpected result %+v", got)
	}
	in.Budget = ""
	got = calculateHypothetical(in)
	if got.Arms[0].Stop == "" || got.Arms[0].Quantity != "" {
		t.Fatalf("missing sizing lost stop %+v", got)
	}
	in.MaxShares = "1.0"
	got = calculateHypothetical(in)
	if got.Errors["max_shares"] == "" {
		t.Fatalf("decimal shares accepted %+v", got)
	}
}

func TestA154DecimalAndRangeBoundaries(t *testing.T) {
	base := hypotheticalInput{Market: "US", Entry: "100", Low: "98", ATR: "2", Tick: "0.01", Min: "1", Max: "200", Budget: "100", MaxShares: "1000", MaxNotional: "100000"}
	for _, s := range []string{"1e2", "1p4", "-1", "0", "1.000000001", "123456789012345678901234567890123", " 100"} {
		in := base
		in.Entry = s
		got := calculateHypothetical(in)
		if got.Errors["entry"] == "" || got.Arms[0].Stop != "" {
			t.Errorf("entry %q accepted: %+v", s, got)
		}
	}
	in := base
	in.Entry = "100.001"
	if calculateHypothetical(in).Errors["entry"] == "" {
		t.Error("off-grid entry accepted")
	}
	in = base
	in.Min = "99.5"
	got := calculateHypothetical(in)
	if got.Arms[0].Stop != "" || got.Arms[0].State != "가격 범위 또는 손절가 오류" {
		t.Errorf("out of range stop accepted: %+v", got)
	}
	in = base
	in.MaxNotional = "150"
	got = calculateHypothetical(in)
	if got.Arms[0].Quantity != "1" || got.Arms[0].Risk != "0.7" {
		t.Errorf("notional cap: %+v", got)
	}
	in = base
	in.Entry = "100000000"
	in.Low = ""
	in.ATR = ""
	in.Tick = "0.00000001"
	in.Min = "1"
	in.Max = "100000001"
	got = calculateHypothetical(in)
	if got.Arms[0].Percent != "0.700000%" {
		t.Errorf("ratio: %+v", got)
	}
}

func TestA154RatioRoundingAndShareCap(t *testing.T) {
	in := hypotheticalInput{Market: "US", Entry: "12800", Low: "12800", ATR: "1", Tick: "1", Min: "1", Max: "20000", Budget: "10000", MaxShares: "1", MaxNotional: "10000000"}
	got := calculateHypothetical(in)
	if got.Arms[1].Percent != "0.007813%" || got.Arms[1].Quantity != "1" || got.Arms[1].Risk != "1" {
		t.Fatalf("rounding/share cap %+v", got.Arms[1])
	}
	in.Entry = "100000000"
	in.Low = "100000000"
	in.Tick = "0.00000001"
	in.Max = "100000001"
	got = calculateHypothetical(in)
	if got.Arms[1].Percent != "<0.000001%" {
		t.Fatalf("tiny ratio %+v", got.Arms[1])
	}
}
