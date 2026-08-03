package protectionreadiness

import "testing"

func FuzzArbitraryAttestationNeverWires(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"schema_version":"protection-readiness/v1"}`))
	f.Add([]byte{0xff, 0x00, 0x01})
	f.Fuzz(func(t *testing.T, data []byte) {
		fixture := newReadinessFixture(t)
		input := fixture.validMarketInput(t, MarketKR, fixture.krPrivate)
		input.File.bytes = append([]byte(nil), data...)
		input.File.Size = int64(len(data))
		input.File.seal = observedFileSeal(input.File)
		result := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: input}))
		if result.Snapshot.Verdict(MarketKR).State == Wired {
			t.Fatalf("arbitrary bytes wired: %q", data)
		}
	})
}

func FuzzSerialMustStrictlyIncrease(f *testing.F) {
	f.Add(uint64(1), uint64(1))
	f.Add(uint64(2), uint64(1))
	f.Add(uint64(1), uint64(2))
	f.Fuzz(func(t *testing.T, accepted, candidate uint64) {
		if accepted == 0 || candidate == 0 {
			return
		}
		fixture := newReadinessFixture(t)
		scope := serialScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}
		fixture.state = newDurableStateWith(readinessNow, map[serialScope]uint64{scope: accepted})
		body := fixture.body(MarketKR)
		body.Serial = candidate
		input := fixture.marketInputForBody(t, body, fixture.krPrivate)
		got := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: input})).Snapshot.Verdict(MarketKR)
		if (got.State == Wired) != (candidate > accepted) {
			t.Fatalf("accepted=%d candidate=%d verdict=%+v", accepted, candidate, got)
		}
	})
}
