package journal

// a200_recognition_verdict_test.go 는 인정 판정 순수 함수(recognitionVerdict — design D2)의 조건 하나하나를 고정함. 원장 시험
// (a200_recognition_test.go)은 생산 경로에서 만들 수 있는 모양만 다루므로, 생산에서 다른 조건이 함께 막아 주는 조건(예: CLOSED
// 이면 거절 유형은 SELL_ON_CLOSED 뿐)도 여기서는 하나씩 깨서 그 조건 자체가 판정을 가르는지 봄 — 우연한 안전에 기대지 않게.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// a200GoodFacts 는 넷이 모두 참인 사실 — 본 사고 모양(31 주, 같은 수량 손절).
func a200GoodFacts() recognitionFacts {
	return recognitionFacts{
		Refusal:         position.RefusalSellOnClosed,
		Adjustments:     []recognitionAdjustment{{Kind: "UNKNOWN", PrevQuantity: "31", NewQuantity: "0"}},
		OriginIntentID:  "i-stop",
		ExitStateFound:  true,
		PendingIntentID: "i-stop",
		OrderAttempts:   []recognitionAttempt{{Kind: "PLACE", State: "CONFIRMED"}},
		IntentQuantity:  "31",
	}
}

func TestA200TheVerdictRecognisesTheIncidentShapeAndRefusesEachBrokenCondition(t *testing.T) {
	if !recognitionVerdict(a200GoodFacts()) {
		t.Fatal("본 사고 모양(넷 모두 참)이 인정되지 않음")
	}
	if convergenceAdjustmentKind != "UNKNOWN" {
		t.Fatalf("수렴이 쓰는 kind = %q, position.Classify(LocalInstance) 는 UNKNOWN", convergenceAdjustmentKind)
	}
	cases := []struct {
		name   string
		mutate func(*recognitionFacts)
		want   bool
	}{
		{"(가) OVERSELL", func(f *recognitionFacts) { f.Refusal = position.RefusalOversell }, false},
		{"(가) UNATTRIBUTED_SELL", func(f *recognitionFacts) { f.Refusal = position.RefusalUnattributedSell }, false},
		{"(가) ENTRY_WHILE_CLOSING", func(f *recognitionFacts) { f.Refusal = position.RefusalEntryWhileClosing }, false},
		{"(나) 조정 행 없음", func(f *recognitionFacts) { f.Adjustments = nil }, false},
		// 둘째 0 행의 prev 를 31 로 둬서 (라)(감소분)가 이 사례를 대신 막지 않게 함 — 「정확히 하나」 자체를 잼(변이 M2).
		{"(나) 0 조정 둘", func(f *recognitionFacts) {
			f.Adjustments = append(f.Adjustments, recognitionAdjustment{Kind: "UNKNOWN", PrevQuantity: "31", NewQuantity: "0"})
		}, false},
		{"(나) 0 조정이 마지막이 아님", func(f *recognitionFacts) {
			f.Adjustments = append(f.Adjustments, recognitionAdjustment{Kind: "EXTERNAL", PrevQuantity: "0", NewQuantity: "5"})
		}, false},
		{"(나) 0 이 아닌 조정만", func(f *recognitionFacts) { f.Adjustments[0].NewQuantity = "11" }, false},
		{"(나) 0 을 decimal 로 적은 0.0", func(f *recognitionFacts) { f.Adjustments[0].NewQuantity = "0.0" }, true},
		{"(나) MANUAL kind", func(f *recognitionFacts) { f.Adjustments[0].Kind = "MANUAL" }, false},
		{"(나) EXTERNAL kind", func(f *recognitionFacts) { f.Adjustments[0].Kind = "EXTERNAL" }, false},
		{"(나) 못 읽는 new_quantity", func(f *recognitionFacts) { f.Adjustments[0].NewQuantity = "x" }, false},
		{"(다) 소유 intent 없음", func(f *recognitionFacts) { f.OriginIntentID = ""; f.PendingIntentID = "" }, false},
		{"(다) exit 상태 없음", func(f *recognitionFacts) { f.ExitStateFound = false }, false},
		{"(다) 발의 비었음(해소됨)", func(f *recognitionFacts) { f.PendingIntentID = "" }, false},
		{"(다) 다른 발의", func(f *recognitionFacts) { f.PendingIntentID = "i-other" }, false},
		{"(라) 시도 없음", func(f *recognitionFacts) { f.OrderAttempts = nil }, false},
		{"(라) PLACE 둘", func(f *recognitionFacts) {
			f.OrderAttempts = append(f.OrderAttempts, recognitionAttempt{Kind: "PLACE", State: "RECORDED"})
		}, false},
		{"(라) PLACE + AMEND", func(f *recognitionFacts) {
			f.OrderAttempts = append(f.OrderAttempts, recognitionAttempt{Kind: "AMEND", State: "CONFIRMED"})
		}, false},
		{"(라) 시도가 AMEND 하나", func(f *recognitionFacts) { f.OrderAttempts[0].Kind = "AMEND" }, false},
		{"(라) 접수 확정 아님", func(f *recognitionFacts) { f.OrderAttempts[0].State = "IN_DOUBT" }, false},
		{"(라) 주문 수량 > 감소분", func(f *recognitionFacts) { f.IntentQuantity = "32" }, false},
		{"(라) 주문 수량 = 감소분(경계)", func(f *recognitionFacts) { f.IntentQuantity = "31" }, true},
		{"(라) 주문 수량 < 감소분", func(f *recognitionFacts) { f.IntentQuantity = "10" }, true},
		{"(라) 주문 수량 0", func(f *recognitionFacts) { f.IntentQuantity = "0" }, false},
		{"(라) 주문 수량 음수", func(f *recognitionFacts) { f.IntentQuantity = "-1" }, false},
		{"(라) 감소분은 prev − new", func(f *recognitionFacts) {
			f.Adjustments[0].PrevQuantity = "21"
			f.IntentQuantity = "31"
		}, false},
		{"(라) 못 읽는 주문 수량", func(f *recognitionFacts) { f.IntentQuantity = "" }, false},
		{"(라) 못 읽는 prev_quantity", func(f *recognitionFacts) { f.Adjustments[0].PrevQuantity = "x" }, false},
	}
	for _, c := range cases {
		f := a200GoodFacts()
		c.mutate(&f)
		if got := recognitionVerdict(f); got != c.want {
			t.Errorf("%s: recognitionVerdict = %v, want %v", c.name, got, c.want)
		}
	}
}
