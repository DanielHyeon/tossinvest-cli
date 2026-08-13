package execgw_test

// a098 R16 — 여섯 번째 차단 트리거는 **자기 사유 코드**를 가진다.
//
// 왜 새 코드인가 (design D8.2): 오늘 진입을 막는 다섯 자리가 전부
// ReasonAlertUndelivered 하나를 쓰고, Notifier.Acknowledge 는 미전달 수가 0 이 되면
// 그 코드의 래치를 푼다. 배달 실행자가 죽은 채 운영자가 밀린 행을 전부 승인하면
// "보낼 주체가 없다"는 차단까지 함께 풀린다 — 보낼 주체는 여전히 없는데.
//
// **골든 파일은 이것을 못 잡는다.** TestReasonCodeEnumIsStable 은
// AllReasonCodes() 를 골든과 비교하는데, 상수를 reason.go 에만 선언하고 열거에
// 안 넣으면 **양쪽 다 그 코드가 없어서 그대로 통과한다.** failclosed.go:283-288 이
// 그 일이 이미 한 번 있었다고 적는다 — 열여섯 개가 그렇게 빠졌었다.
//
// 그래서 여기서 네 가지를 따로 잰다: 길이 · 명시적 소속 · 문자열 · latchOrder 소속.

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// a098ReasonCodeCountBeforeSenderDown is what AllReasonCodes() enumerated before
// this change. Measured, not remembered: `wc -l reason_codes.golden` = 29 on
// base e6c4636a.
const a098ReasonCodeCountBeforeSenderDown = 29

// TestTheSenderDownReasonIsRegisteredInTheEnumeration is check ① and ②.
//
// Length and membership are separate assertions on purpose. Membership alone
// passes if somebody registers the code twice and drops another; length alone
// passes if they register any code at all.
func TestTheSenderDownReasonIsRegisteredInTheEnumeration(t *testing.T) {
	codes := execgw.AllReasonCodes()

	if got, want := len(codes), a098ReasonCodeCountBeforeSenderDown+1; got != want {
		t.Errorf("AllReasonCodes() has %d codes, want %d — exactly one more than before a098", got, want)
	}

	found := false
	for _, c := range codes {
		if c == execgw.ReasonAlertSenderDown {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AllReasonCodes() does not contain ReasonAlertSenderDown; "+
			"declaring the constant in reason.go is not registering it (see failclosed.go:283-288). got %v", codes)
	}
}

// TestTheSenderDownReasonIsNotTheUndeliveredReason is why the code exists at all.
//
// If the two were the same string, an operator acknowledging the backlog would
// clear the "nobody is sending" latch as a side effect (notifier.go:507-512).
func TestTheSenderDownReasonIsNotTheUndeliveredReason(t *testing.T) {
	if execgw.ReasonAlertSenderDown == execgw.ReasonAlertUndelivered {
		t.Fatal("ReasonAlertSenderDown equals ReasonAlertUndelivered — " +
			"Acknowledge clears the undelivered latch when the backlog empties, " +
			"which would release the sender-down block while the sender is still dead")
	}
	if got, want := string(execgw.ReasonAlertSenderDown), "critical_alert_sender_down"; got != want {
		t.Errorf("ReasonAlertSenderDown = %q, want %q — the string lands in the journal, so it is a contract", got, want)
	}
}

// TestTheSenderDownReasonHasADeterministicLatchPrecedence is check ④.
//
// checkAccountEntryLocked walks latchOrder first (retry.go:540-544) and only then
// falls back to ranging the latch map (`:545-547`). A code missing from latchOrder
// still blocks — that is why forgetting it is invisible — but when two such codes
// are latched together the reported one becomes map-iteration order, which is
// randomised per range in Go.
//
// So the observation pairs sender-down with a code that is *also* absent from
// latchOrder (ReasonGuardianMissing). If sender-down is registered the answer is
// stable across every call; if it is not, both fall through to the map and the
// answer flaps.
func TestTheSenderDownReasonHasADeterministicLatchPrecedence(t *testing.T) {
	gate := execgw.NewEntryGate(clock.System(), map[execgw.RequiredQuery]time.Duration{})
	gate.Block(execgw.ReasonAlertSenderDown, "delivery executor stopped")
	gate.Block(execgw.ReasonGuardianMissing, "no guardian decision")

	const rounds = 200
	for i := range rounds {
		rejected := gate.CheckEntry()
		if rejected == nil {
			t.Fatalf("round %d: gate with two latches allowed entry", i)
		}
		if rejected.Reason != execgw.ReasonAlertSenderDown {
			t.Fatalf("round %d: CheckEntry reported %q, want %q — "+
				"a reason missing from latchOrder still blocks, but which one is reported "+
				"becomes map-iteration order (retry.go:545-547)",
				i, rejected.Reason, execgw.ReasonAlertSenderDown)
		}
	}
}
