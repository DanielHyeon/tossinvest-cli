package riskbucket

// a066 6.1 (B)(2) — Branch Test Map 에서 "NOT covered" 였던 ApplyFill·recomputeOverageLatches 분기 중 입력으로 닿는
// 것을 시험으로 채움. 모든 거절은 상태를 바꾸지 않아야 함(ApplyFill 계약: 오류면 손대지 않은 사본).
// 닿지 않는 분기(검증이 앞에서 같은 값을 먼저 해석함)는 review.md 6.1 절에 걸은 범위와 함께 적음.

import (
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

// refusedUnchanged 는 거절 사유와 "상태 불변"을 함께 단언함.
func refusedUnchanged(t *testing.T, name string, state FillState, event FillEvent, reason string) {
	t.Helper()
	original := cloneFillStateForTest(state)
	next, _, err := ApplyFill(state, event)
	var refused *RefusalError
	if !errors.As(err, &refused) || refused.Code != RefusalFillEvidenceInconsistent || refused.Field != reason {
		t.Fatalf("%s: err=%v, want FILL_EVIDENCE_INCONSISTENT/%s", name, err, reason)
	}
	if !reflect.DeepEqual(state, original) || !reflect.DeepEqual(next, original) {
		t.Fatalf("%s: refused fill mutated state", name)
	}
}

// unknownFirst 는 실제 증거 없이 한 번 적용한 상태(UNKNOWN_ACTUAL latch)와 그 사건을 돌려줌 — 중복 완성 경로의 출발점.
func unknownFirst(t *testing.T) (FillState, FillEvent) {
	t.Helper()
	state, event := fillFixture("100", "50")
	event.NewCumulativeFill = 4
	next, result, err := ApplyFill(state, event)
	if err != nil || result.DeltaQuantity != 4 {
		t.Fatalf("unknown first fill: result=%+v err=%v", result, err)
	}
	return next, event
}

func TestA066ApplyFillTargetHeldShapeIsRefused(t *testing.T) {
	// B4: target HELD 의 bucket 수가 예약 bucket 수와 다름.
	state, event := fillFixture("100", "50")
	event.TargetHeldMinor = map[BucketKey]string{}
	for key := range event.ReservedMinor {
		event.TargetHeldMinor[key] = "50"
		break
	}
	refusedUnchanged(t, "target count", state, event, "target_held_bucket_count")

	// B6: target HELD 값이 해석되지 않음.
	state, event = fillFixture("100", "50")
	event.TargetHeldMinor = map[BucketKey]string{}
	for key := range event.ReservedMinor {
		event.TargetHeldMinor[key] = "50"
	}
	for key := range event.TargetHeldMinor {
		event.TargetHeldMinor[key] = "-1"
		break
	}
	refusedUnchanged(t, "target unparseable", state, event, "target_held_usage")
}

func TestA066ApplyFillSameFillIDWithAnotherWatermarkIsRefused(t *testing.T) {
	// B13: 같은 fill id 가 다른 누적 체결량으로 다시 옴.
	state, event := unknownFirst(t)
	event.NewCumulativeFill = 5
	refusedUnchanged(t, "watermark", state, event, "fill_identity_watermark")
}

func TestA066ApplyFillRetryWithStillUnknownActualIsADuplicateNoOp(t *testing.T) {
	// B15: 첫 적용이 UNKNOWN 이었고 재시도도 실제 증거가 없음 → 중복, 상태 불변(latch 유지).
	state, event := unknownFirst(t)
	original := cloneFillStateForTest(state)
	next, result, err := ApplyFill(state, event)
	if err != nil || !result.Duplicate || result.ActualEvidenceCompleted || !reflect.DeepEqual(next, original) {
		t.Fatalf("still-unknown retry: result=%+v err=%v changed=%v", result, err, !reflect.DeepEqual(next, original))
	}
	if !next.OwnerLatches[LatchUnknownActualRisk] {
		t.Fatal("still-unknown retry cleared the UNKNOWN_ACTUAL latch")
	}
}

func TestA066ApplyFillCorruptStoredFillRecordIsRefusedOnCompletion(t *testing.T) {
	for _, tc := range []struct {
		name, field, reason string
	}{
		{"B17 transfer", "transfer", "record_transfer"},
		{"B18 filled", "filled", "record_filled"},
	} {
		state, event := unknownFirst(t)
		order := state.Orders[event.OrderID]
		record := order.Fills[event.FillID]
		for key := range record.TransferMinor {
			if tc.field == "transfer" {
				record.TransferMinor[key] = "x"
			} else {
				record.FilledMinor[key] = "x"
			}
		}
		order.Fills[event.FillID] = record
		state.Orders[event.OrderID] = order
		event.Actual = actualEvidence("12", "1", "0")
		refusedUnchanged(t, tc.name, state, event, tc.reason)
	}
}

func TestA066ApplyFillCompletionOverflowIsRefused(t *testing.T) {
	// B22: 중복 완성에서 filled 사용량 + 차이가 256비트를 넘음.
	state, event := unknownFirst(t)
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	for key, usage := range state.Buckets {
		usage.FilledMinor = max256
		state.Buckets[key] = usage
	}
	event.Actual = actualEvidence("12", "1", "0")
	refusedUnchanged(t, "completion overflow", state, event, "filled_usage_overflow")
}

func TestA066ApplyFillCorruptSharedUsageIsRefusedOnBothPaths(t *testing.T) {
	// B38(첫 적용) · B23(중복 완성) → recomputeOverageLatches B7: 다른 owner 의 공유 사용량이 해석되지 않음.
	state, event := fillFixture("100", "50")
	state.SharedUsedMinor = map[BucketKey]string{}
	for key := range state.Buckets {
		state.SharedUsedMinor[key] = "corrupt"
		break
	}
	refusedUnchanged(t, "first apply", state, event, "overage_shared_usage")

	state, event = unknownFirst(t)
	state.SharedUsedMinor = map[BucketKey]string{}
	for key := range state.Buckets {
		state.SharedUsedMinor[key] = "corrupt"
		break
	}
	event.Actual = actualEvidence("12", "1", "0")
	refusedUnchanged(t, "completion", state, event, "overage_shared_usage")
}

func TestA066ApplyFillUsageOverflowInOverageIsRefused(t *testing.T) {
	// recomputeOverageLatches B5: filled + held 가 256비트를 넘음(각각은 유효).
	half := new(big.Int).Lsh(big.NewInt(1), 255).String()
	state, event := fillFixture(half, half)
	for key, usage := range state.Buckets {
		usage.FilledMinor = half
		state.Buckets[key] = usage
	}
	refusedUnchanged(t, "usage overflow", state, event, "overage_usage_overflow")

	// B8: 이 owner 의 합은 유효하지만 공유 사용량을 더하면 넘음.
	state, event = fillFixture("100", "50")
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	state.SharedUsedMinor = map[BucketKey]string{}
	for key := range state.Buckets {
		state.SharedUsedMinor[key] = max256
	}
	refusedUnchanged(t, "shared overflow", state, event, "overage_shared_usage_overflow")
}

func TestA066ApplyFillNonIncreasingCumulativeIsRefused(t *testing.T) {
	// B24: 새 fill id 인데 누적 체결량이 늘지 않음.
	state, event := unknownFirst(t)
	event.FillID = "fill-2"
	refusedUnchanged(t, "cumulative", state, event, "cumulative_fill")
}

func TestA066ApplyFillHeldDeductionIsCappedAtHeldAndLatchesOverage(t *testing.T) {
	// B34: 옮길 양이 남은 HELD 보다 큼(예약 50, HELD 10) → HELD 는 0 에서 멈추고 RISK_OVERAGE latch.
	state, event := fillFixture("1000", "50")
	for key, usage := range state.Buckets {
		usage.HeldMinor = "10"
		state.Buckets[key] = usage
	}
	event.NewCumulativeFill = event.OrderQuantity
	next, _, err := ApplyFill(state, event)
	if err != nil {
		t.Fatal(err)
	}
	for key, usage := range next.Buckets {
		if usage.HeldMinor != "0" || !usage.Latches[LatchRiskOverage] {
			t.Fatalf("%s: held=%s overage latch=%v, want 0 and latched", key.Dimension, usage.HeldMinor, usage.Latches[LatchRiskOverage])
		}
	}
	if !next.OwnerLatches[LatchRiskOverage] {
		t.Fatal("owner RISK_OVERAGE latch not set")
	}
}

// TestA066ApplyFillValidatesBucketShapeBeforeAnyArithmetic 는 ApplyFill B2(validateFillBuckets)가 저장 상태·예약의 모양을
// **산술 전에** 거절함을 사유 칸으로 고정함. BTM 행 변이에서 B2 를 꺼도 초록이었음 — 뒤의 해석 가드(overage_limit 등)가
// 같은 입력을 다른 사유로 막았기 때문. 사유 칸이 validateFillBuckets 의 것이어야 통과함.
func TestA066ApplyFillValidatesBucketShapeBeforeAnyArithmetic(t *testing.T) {
	anyKey := func(m map[BucketKey]string) BucketKey {
		for key := range m {
			return key
		}
		return BucketKey{}
	}
	for _, tc := range []struct {
		name, reason string
		corrupt      func(*FillState, *FillEvent)
	}{
		{"bucket count", "bucket_count", func(s *FillState, e *FillEvent) { delete(s.Buckets, anyKey(e.ReservedMinor)) }},
		{"reservation outside the state", "reservation_bucket", func(s *FillState, e *FillEvent) {
			key := anyKey(e.ReservedMinor)
			amount := e.ReservedMinor[key]
			delete(e.ReservedMinor, key)
			key.Value += "-other"
			e.ReservedMinor[key] = amount
		}},
		{"unparseable stored limit", "", func(s *FillState, e *FillEvent) {
			key := anyKey(e.ReservedMinor)
			usage := s.Buckets[key]
			usage.LimitMinor = "corrupt"
			s.Buckets[key] = usage
		}},
	} {
		state, event := fillFixture("100", "50")
		tc.corrupt(&state, &event)
		original := cloneFillStateForTest(state)
		next, _, err := ApplyFill(state, event)
		var refused *RefusalError
		if !errors.As(err, &refused) || refused.Code != RefusalFillEvidenceInconsistent {
			t.Fatalf("%s: err=%v", tc.name, err)
		}
		want := tc.reason
		if want == "" {
			// 사유 칸은 "<차원>_limit" — 뒤의 recomputeOverageLatches 는 "overage_limit" 을 씀.
			if !strings.HasSuffix(refused.Field, "_limit") || refused.Field == "overage_limit" {
				t.Fatalf("%s: refused by %q, want validateFillBuckets' <dimension>_limit", tc.name, refused.Field)
			}
		} else if refused.Field != want {
			t.Fatalf("%s: refused by %q, want %q", tc.name, refused.Field, want)
		}
		if !reflect.DeepEqual(next, original) {
			t.Fatalf("%s: refused fill mutated state", tc.name)
		}
	}
}

// TestA066ApplyFillRecordKeepsTheActualEvidenceItWasValuedWith 는 ApplyFill B25 를 고정함: 첫 적용에서 실제 체결 증거가
// 있으면 기록은 그 증거의 사본을 가짐(입력과 같은 값, 다른 포인터). BTM 행 변이에서 이 대입을 꺼도 초록이었음 — 생산
// 코드에 이 필드를 읽는 곳이 없음(git grep internal/ 비시험: fill.go 의 대입·복제뿐). 그래서 이것은 riskbucket 순수 함수의
// 반환 계약으로만 서 있고, 그 계약을 여기서 잼.
func TestA066ApplyFillRecordKeepsTheActualEvidenceItWasValuedWith(t *testing.T) {
	state, event := fillFixture("100", "50")
	event.NewCumulativeFill = 4
	event.Actual = actualEvidence("12", "1", "0")
	next, _, err := ApplyFill(state, event)
	if err != nil {
		t.Fatal(err)
	}
	record := next.Orders[event.OrderID].Fills[event.FillID]
	if !record.ActualKnown || record.ActualEvidence == nil || !reflect.DeepEqual(*record.ActualEvidence, *event.Actual) || record.ActualEvidence == event.Actual {
		t.Fatalf("record evidence=%+v known=%v, want a copy of the valued evidence", record.ActualEvidence, record.ActualKnown)
	}
}

// TestA066ApplyFillFirstApplyOverflowIsRefusedAtTheAddition 는 첫 적용 경로의 filled 합 넘침(B36)이 **덧셈 자리에서**
// 거절됨을 사유 칸으로 고정함. BTM 행 변이에서 B36 을 꺼도 초록이었음 — nil 합이 문자열 "<nil>" 로 저장된 뒤
// recomputeOverageLatches 의 해석(overage_filled)이 대신 막았기 때문. 사유는 "filled_usage_overflow" 여야 함.
func TestA066ApplyFillFirstApplyOverflowIsRefusedAtTheAddition(t *testing.T) {
	state, event := fillFixture(new(big.Int).Lsh(big.NewInt(1), 255).String(), "1")
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	for key, usage := range state.Buckets {
		usage.FilledMinor = max256
		state.Buckets[key] = usage
	}
	event.NewCumulativeFill = event.OrderQuantity
	refusedUnchanged(t, "first-apply overflow", state, event, "filled_usage_overflow")
}
