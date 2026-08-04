package riskbucket

import (
	"math/big"
	"reflect"
	"testing"
	"time"
)

func TestApplyFillTransfersHeldAndUsesGreaterActualExposure(t *testing.T) {
	state, event := fillFixture("100", "50")
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("12", "1", "0")

	next, result, err := ApplyFill(state, event)
	if err != nil {
		t.Fatalf("ApplyFill: %v", err)
	}
	if result.Duplicate || result.DeltaQuantity != 4 {
		t.Fatalf("result = %#v", result)
	}
	for key, usage := range next.Buckets {
		if usage.HeldMinor != "30" || usage.FilledMinor != "48" {
			t.Fatalf("%s usage = held %s filled %s, want 30/48", key.Dimension, usage.HeldMinor, usage.FilledMinor)
		}
	}
	if !reflect.DeepEqual(state, mustOriginalFillState("100", "50")) {
		t.Fatal("ApplyFill mutated its input state")
	}
}

func TestApplyFillRetryIsIdempotent(t *testing.T) {
	state, event := fillFixture("100", "50")
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("12", "1", "0")
	first, _, err := ApplyFill(state, event)
	if err != nil {
		t.Fatal(err)
	}
	second, result, err := ApplyFill(first, event)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Duplicate || !reflect.DeepEqual(first, second) {
		t.Fatalf("retry changed state: duplicate=%v", result.Duplicate)
	}
}

func TestApplyFillUsesOrderKeyToKeepDecisionOrdersIndependent(t *testing.T) {
	state, first := fillFixture("200", "100")
	first.OrderKey = "decision-1/order"
	first.OrderID = "broker-reused"
	first.NewCumulativeFill = 2
	next, result, err := ApplyFill(state, first)
	if err != nil || result.DeltaQuantity != 2 {
		t.Fatalf("first=%+v err=%v", result, err)
	}
	second := first
	second.FillID = "fill-decision-2"
	second.OrderKey = "decision-2/order"
	second.NewCumulativeFill = 2
	next, result, err = ApplyFill(next, second)
	if err != nil || result.DeltaQuantity != 2 {
		t.Fatalf("second=%+v err=%v", result, err)
	}
	if len(next.Orders) != 2 || next.Orders[first.OrderKey].CumulativeFill != 2 || next.Orders[second.OrderKey].CumulativeFill != 2 {
		t.Fatalf("orders=%+v", next.Orders)
	}
	for key, usage := range next.Buckets {
		if usage.HeldMinor != "60" || usage.FilledMinor != "40" {
			t.Fatalf("%+v usage=%+v", key, usage)
		}
	}
}

func TestApplyFillUnknownActualLatchesThenCompletesMonotonically(t *testing.T) {
	state, event := fillFixture("100", "50")
	event.NewCumulativeFill = 4

	unknown, _, err := ApplyFill(state, event)
	if err != nil {
		t.Fatal(err)
	}
	for key, usage := range unknown.Buckets {
		if usage.FilledMinor != "20" || !usage.Latches[LatchUnknownActualRisk] {
			t.Fatalf("%s unknown usage = %#v", key.Dimension, usage)
		}
	}
	if !unknown.OwnerLatches[LatchUnknownActualRisk] || !unknown.EntryBlocked() {
		t.Fatal("unknown actual did not block new exposure")
	}

	event.Actual = actualEvidence("12", "1", "0")
	completed, result, err := ApplyFill(unknown, event)
	if err != nil {
		t.Fatal(err)
	}
	if result.Duplicate || !result.ActualEvidenceCompleted {
		t.Fatalf("completion result = %#v", result)
	}
	for key, usage := range completed.Buckets {
		if usage.FilledMinor != "48" || usage.Latches[LatchUnknownActualRisk] {
			t.Fatalf("%s completed usage = %#v", key.Dimension, usage)
		}
	}
	if completed.OwnerLatches[LatchUnknownActualRisk] {
		t.Fatal("owner unknown latch remained after all evidence completed")
	}
}

func TestApplyFillLatchesEveryBucketOnOverageWithoutDroppingFill(t *testing.T) {
	state, event := fillFixture("40", "50")
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("12", "1", "0")

	next, result, err := ApplyFill(state, event)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeltaQuantity != 4 || next.Orders[event.OrderID].CumulativeFill != 4 {
		t.Fatalf("authoritative fill watermark was not preserved: %#v", result)
	}
	for key, usage := range next.Buckets {
		if usage.OverageMinor != "38" || !usage.Latches[LatchRiskOverage] {
			t.Fatalf("%s overage usage = %#v", key.Dimension, usage)
		}
	}
	if !next.OwnerLatches[LatchRiskOverage] || !next.EntryBlocked() {
		t.Fatal("overage did not latch owner entry")
	}
}

func TestApplyLateFillAfterHeldReleasePreservesWatermarkAndLatches(t *testing.T) {
	state, event := fillFixture("40", "0")
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("12", "1", "0")

	next, result, err := ApplyFill(state, event)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeltaQuantity != 4 || next.Orders[event.OrderID].CumulativeFill != 4 {
		t.Fatalf("late fill was dropped: %#v", result)
	}
	for key, usage := range next.Buckets {
		if usage.HeldMinor != "0" || usage.FilledMinor != "48" || !usage.Latches[LatchRiskOverage] {
			t.Fatalf("%s late-fill usage = %#v", key.Dimension, usage)
		}
	}
}

func TestApplyFillMalformedActualBecomesUnknownWithoutDroppingFill(t *testing.T) {
	state, event := fillFixture("100", "50")
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("not-a-price", "1", "0")

	next, result, err := ApplyFill(state, event)
	if err != nil {
		t.Fatalf("authoritative fill must survive unknown actual: %v", err)
	}
	if result.DeltaQuantity != 4 || next.Orders[event.OrderID].CumulativeFill != 4 {
		t.Fatal("unknown actual dropped fill watermark")
	}
	for key, usage := range next.Buckets {
		if usage.FilledMinor != "20" || !usage.Latches[LatchUnknownActualRisk] {
			t.Fatalf("%s malformed-actual usage = %#v", key.Dimension, usage)
		}
	}
}

func TestApplyFillPartialThenReplacementCumulativeUsesProportionalDelta(t *testing.T) {
	state, first := fillFixture("200", "50")
	first.FillID = "fill-1"
	first.NewCumulativeFill = 4
	first.Actual = actualEvidence("12", "1", "0")
	afterFirst, _, err := ApplyFill(state, first)
	if err != nil {
		t.Fatal(err)
	}

	second := first
	second.FillID = "fill-2"
	second.NewCumulativeFill = 7
	afterSecond, result, err := ApplyFill(afterFirst, second)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeltaQuantity != 3 {
		t.Fatalf("delta quantity = %d, want 3", result.DeltaQuantity)
	}
	for key, usage := range afterSecond.Buckets {
		if usage.HeldMinor != "15" || usage.FilledMinor != "84" {
			t.Fatalf("%s usage = held %s filled %s, want 15/84", key.Dimension, usage.HeldMinor, usage.FilledMinor)
		}
	}
}

func TestApplyFillErrorIsCrashPure(t *testing.T) {
	state, event := fillFixture("100", "50")
	for key := range event.ReservedMinor {
		event.ReservedMinor[key] = "not-a-minor-amount"
		break
	}
	original := cloneFillStateForTest(state)
	next, _, err := ApplyFill(state, event)
	if err == nil {
		t.Fatal("expected invalid persisted reservation error")
	}
	if !reflect.DeepEqual(state, original) || !reflect.DeepEqual(next, original) {
		t.Fatal("failed fill transition was not atomic/pure")
	}
}

func TestApplyFillCorruptPersistedUsageIsCrashPure(t *testing.T) {
	state, event := fillFixture("100", "50")
	for key, usage := range state.Buckets {
		usage.LimitMinor = "corrupt"
		state.Buckets[key] = usage
		break
	}
	original := cloneFillStateForTest(state)
	next, _, err := ApplyFill(state, event)
	if err == nil {
		t.Fatal("expected corrupt persisted usage error")
	}
	if !reflect.DeepEqual(state, original) || !reflect.DeepEqual(next, original) {
		t.Fatal("corrupt-state failure was not atomic/pure")
	}
}

func TestApplyFillRejectsStoredMinorOverflowWithoutPartialMutation(t *testing.T) {
	state, event := fillFixture(new(big.Int).Lsh(big.NewInt(1), 255).String(), "1")
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	for key, usage := range state.Buckets {
		usage.FilledMinor = max256
		state.Buckets[key] = usage
	}
	event.NewCumulativeFill = event.OrderQuantity
	original := cloneFillStateForTest(state)

	next, _, err := ApplyFill(state, event)
	if !IsRefusal(err, RefusalFillEvidenceInconsistent) {
		t.Fatalf("overflow error = %v", err)
	}
	if !reflect.DeepEqual(state, original) || !reflect.DeepEqual(next, original) {
		t.Fatal("overflow failure partially mutated bucket or fill watermark state")
	}
}

func TestApplyFillRejectsMismatchedActualFXCurrencyPairAsUnknown(t *testing.T) {
	state, event := fillFixture("100", "50")
	event.QuoteCurrency = "USD"
	event.BaseCurrency = "KRW"
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("12", "1300", "0")
	event.Actual.QuoteCurrency = "USD"
	event.Actual.BaseCurrency = "USD"

	next, result, err := ApplyFill(state, event)
	if err != nil {
		t.Fatalf("authoritative fill must survive mismatched FX evidence: %v", err)
	}
	if result.DeltaQuantity != 4 || next.Orders[event.OrderID].CumulativeFill != 4 {
		t.Fatal("mismatched FX pair dropped authoritative fill")
	}
	for key, usage := range next.Buckets {
		if usage.FilledMinor != "20" || !usage.Latches[LatchUnknownActualRisk] {
			t.Fatalf("%s mismatched-FX usage = %#v", key.Dimension, usage)
		}
	}
}

func TestEntryBlockedConsultsBucketLatchesEvenWithoutOwnerAggregate(t *testing.T) {
	key := BucketKey{Dimension: DimensionSymbol, Value: "005930", PolicyVersion: "v1"}
	state := FillState{Buckets: map[BucketKey]BucketUsage{key: {Latches: map[Latch]bool{LatchRiskOverage: true}}}}
	if !state.EntryBlocked() {
		t.Fatal("bucket risk-overage latch did not block entry")
	}
	state.Buckets[key] = BucketUsage{Latches: map[Latch]bool{LatchUnknownActualRisk: true}}
	if !state.EntryBlocked() {
		t.Fatal("bucket unknown-actual latch did not block entry")
	}
}

func FuzzApplyFillRetryIsPure(f *testing.F) {
	for _, cumulative := range []uint64{1, 4, 10} {
		f.Add(cumulative)
	}
	f.Fuzz(func(t *testing.T, cumulative uint64) {
		if cumulative == 0 || cumulative > 10 {
			t.Skip()
		}
		state, event := fillFixture("1000", "100")
		event.NewCumulativeFill = cumulative
		event.Actual = actualEvidence("9", "1", "0")
		first, _, err := ApplyFill(state, event)
		if err != nil {
			t.Fatal(err)
		}
		second, result, err := ApplyFill(first, event)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Duplicate || !reflect.DeepEqual(first, second) {
			t.Fatal("retry was not idempotent")
		}
	})
}

func fillFixture(limit, held string) (FillState, FillEvent) {
	buckets := testBuckets(limit)
	state := FillState{Buckets: make(map[BucketKey]BucketUsage), Orders: make(map[string]OrderFillState), OwnerLatches: make(map[Latch]bool)}
	reserved := make(map[BucketKey]string)
	for _, bucket := range buckets {
		state.Buckets[bucket.Key] = BucketUsage{LimitMinor: limit, HeldMinor: held, FilledMinor: "0", OverageMinor: "0", Latches: make(map[Latch]bool)}
		reserved[bucket.Key] = held
	}
	event := FillEvent{FillID: "fill-1", OrderID: "order-1", OrderQuantity: 10, NewCumulativeFill: 1, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW"}
	return state, event
}

func mustOriginalFillState(limit, held string) FillState {
	state, _ := fillFixture(limit, held)
	return state
}

func actualEvidence(price, fx, fee string) *ActualFillEvidence {
	return &ActualFillEvidence{
		QuoteCurrency:         "KRW",
		BaseCurrency:          "KRW",
		PriceQuote:            price,
		FXRateQuoteToBase:     fx,
		AllocatedFeeBaseMinor: fee,
		Price:                 Evidence{Source: "official-fill", Version: "v1", Digest: "fill-price", Official: true, Frozen: true, ObservedAt: testNow, FreshUntil: testNow.Add(time.Minute)},
		FX:                    Evidence{Source: "official-fx", Version: "v1", Digest: "fill-fx", Official: true, Frozen: true, ObservedAt: testNow, FreshUntil: testNow.Add(time.Minute)},
		EvaluatedAt:           testNow,
	}
}

func cloneFillStateForTest(in FillState) FillState {
	return cloneFillState(in)
}
