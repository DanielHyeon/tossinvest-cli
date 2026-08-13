package engine

// a098 R8 — 재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다.
//
// EntryGate.Block 은 프로세스 메모리 안의 맵뿐이다 (execgw/retry.go:498-505).
// 원장에는 미전달 critical 행이 남아 있는데 새 프로세스는 그것을 모른 채 진입을 연다.
// 경보가 아직 아무에게도 안 갔다는 사실은 재시작이 바꾸지 않으므로, 그 래치는
// 원장에서 다시 세워져야 한다.
//
// 저장소에 **같은 결함을 이미 고친 선례**가 있다: gateway.go:216-226 의
// tracker.Restore 가 RECONCILE 계열 래치를 원장에서 재건하고, 그 주석이
// *"without this call a restart silently clears every block a disagreement raised"*
// 라고 적는다. 알림 래치만 그 복원을 안 받고 있다.
//
// ⚠ 관측을 CheckEntry 로 쓰면 안 된다. 기동 직후의 게이트는 기본 staleness 임계를
// 들고 있어서 **한 번도 관측 안 된 질의 때문에 이미 진입을 막는다**
// (gateway.go:211-213). 그래서 CheckEntry 는 4.6 이 있든 없든 non-nil 이고,
// 그것으로는 born-GREEN 이다. R8 이 Blocks() 를 지정하는 이유가 이것이다 —
// 래치 맵은 오직 Block 이 부른 것만 담는다.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// a098BuildGateway assembles the production execution stack the way
// engine.go:490 does, with the two collaborators that assembly dereferences.
//
// The official client is constructed but never called: buildGateway reads its
// BaseURL and takes AuthHeaders as a value (gateway.go:265-267). Passing nil
// panics there, which is a wiring fact worth stating rather than a reason to
// test somewhere else — the point of R8 is the *production* assembly.
//
// ⛔ The trading service is built on a **nil broker** on purpose. execgw.New
// requires a Service to exist ("execgw: a trading service is required") but this
// test never submits anything, and a service with no broker cannot: every
// mutation path dereferences it. That is the mechanical guarantee, not a comment
// — an automated test must be unable to reach a live account, not merely
// uninterested in one.
func a098BuildGateway(t *testing.T, ctx context.Context, j *journal.Journal) engineWiring {
	t.Helper()
	dir := t.TempDir()
	wiring, err := buildGateway(ctx, gatewayInputs{
		journal:    j,
		accountRef: "123-45",
		clock:      clock.System(),
		configDir:  dir,
		trading:    trading.NewService(config.Trading{}, nil),
		official: official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(dir, "token.json"),
		),
	})
	if err != nil {
		t.Fatalf("buildGateway: %v", err)
	}
	return wiring
}

// a098SeedUndeliveredCriticalAlert leaves exactly what a dead process leaves
// behind: a PENDING critical row nobody has delivered.
func a098SeedUndeliveredCriticalAlert(t *testing.T, j *journal.Journal) {
	t.Helper()
	ctx := context.Background()
	if _, err := j.EnqueueAlert(ctx, journal.Alert{
		EventKey: "a098-r8-undelivered",
		Type:     "engine.exit_proposal_refused",
		Severity: "critical",
		Title:    "an exit proposal was refused",
		Body:     "nobody has seen this",
	}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	n, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if n != 1 {
		t.Fatalf("seeded backlog = %d undelivered, want 1", n)
	}
}

// TestARestartWithAnUndeliveredCriticalAlertKeepsEntryLatched is the whole of R8.
//
// The observation is on the wiring's gate the moment assembly returns — that is
// before any loop has started, and therefore before any entry decision. Asserting
// only the final state would pass an implementation that latches on the delivery
// loop's first tick, which is after the loops are running and lets an order
// through in between (6라운드 P3).
func TestARestartWithAnUndeliveredCriticalAlertKeepsEntryLatched(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	a098SeedUndeliveredCriticalAlert(t, j)

	wiring := a098BuildGateway(t, ctx, j)

	blocks := wiring.entry.Blocks()
	detail, latched := blocks[execgw.ReasonAlertUndelivered]
	if !latched {
		t.Fatalf("the entry gate has no %q latch after assembly, so a restart released it: "+
			"the ledger still holds an undelivered critical alert and the new process opened entry anyway. blocks = %v",
			execgw.ReasonAlertUndelivered, blocks)
	}
	if detail == "" {
		t.Error("the latch has no detail; an operator reading Blocks() needs to know what is undelivered")
	}
}

// TestAnEmptyOutboxLeavesTheGateUnlatchedOnRestart is the other half, and it is
// what keeps the restore from being "always block".
//
// 사용자 결정 5-1: 복원은 **잠그는 쪽으로만** 한다. 미전달이 0이면 아무것도 안 한다 —
// 그런데 "아무것도 안 한다"는 "Clear 를 부른다"와 다르다. 승인된 정본
// (openspec/specs/engine-safety/spec.md:147-152)이 전달 복구 뒤 **수동 확인**을
// 요구하므로, 복원이 래치를 **푸는** 일은 절대 없어야 한다.
func TestAnEmptyOutboxLeavesTheGateUnlatchedOnRestart(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}

	wiring := a098BuildGateway(t, ctx, j)

	if _, latched := wiring.entry.Blocks()[execgw.ReasonAlertUndelivered]; latched {
		t.Errorf("an empty outbox latched the gate anyway; the restore must read the ledger, not assume")
	}
}

// TestTheRestoreNeverClearsALatchItDidNotSet pins the direction of 결정 5-1.
//
// A restore written as "derive the latch from the undelivered count" would clear
// a latch some other condition raised, the moment the backlog empties. That is
// less conservative than the approved requirement, which asks for a manual
// confirmation after delivery recovers (engine-safety spec.md:147-152).
//
// ⛔ 이 테스트의 첫 판은 **이빨이 없었다.** buildGateway 로 조립한 뒤에 Block 을
// 걸었는데, 복원은 그보다 **먼저** 끝나 있다 — 그래서 Clear 를 부르는 구현을 두고
// 돌려도 통과했다(뮤테이션으로 확인). 관측하려는 순서가 코드의 순서와 반대였다.
//
// 그래서 leaf 를 직접 부른다. 게이트를 **미리 잠근 채** 복원을 돌리는 것이
// "안 푼다"를 관측할 수 있는 유일한 순서다.
func TestTheRestoreNeverClearsALatchItDidNotSet(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t) // the outbox is empty
	gate := execgw.NewEntryGate(clock.System(), nil)
	gate.Block(execgw.ReasonAlertUndelivered, "raised by something else")

	if err := restoreAlertEntryLatch(ctx, j, gate); err != nil {
		t.Fatalf("restoreAlertEntryLatch: %v", err)
	}

	if _, latched := gate.Blocks()[execgw.ReasonAlertUndelivered]; !latched {
		t.Error("the latch disappeared while the outbox was empty; " +
			"the restore latches only, it never clears — recovering delivery is not somebody having read the alert")
	}
}

// TestTheRestoredLatchDetailCarriesNoAlertContent is 안전 불변식 8 at this seam.
//
// The gate's detail is read wherever the gate's status is read, and the rows the
// restore counted carry the alert title, body, payload and the account they
// concern (alertSelect, outbox.go:510-513). Counting is all it may report.
func TestTheRestoredLatchDetailCarriesNoAlertContent(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if _, err := j.EnqueueAlert(ctx, journal.Alert{
		EventKey: "a098-r8-secretish",
		Type:     "engine.exit_proposal_refused",
		Severity: "critical",
		Title:    "SECRET-TITLE-9f3a",
		Body:     "SECRET-BODY-9f3a",
		Payload:  `{"account":"SECRET-ACCOUNT-9f3a"}`,
	}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}

	gate := execgw.NewEntryGate(clock.System(), nil)
	if err := restoreAlertEntryLatch(ctx, j, gate); err != nil {
		t.Fatalf("restoreAlertEntryLatch: %v", err)
	}

	detail := gate.Blocks()[execgw.ReasonAlertUndelivered]
	if detail == "" {
		t.Fatal("no latch was set for a non-empty outbox")
	}
	for _, secret := range []string{"SECRET-TITLE-9f3a", "SECRET-BODY-9f3a", "SECRET-ACCOUNT-9f3a"} {
		if strings.Contains(detail, secret) {
			t.Errorf("the latch detail leaks %q: %q", secret, detail)
		}
	}
	if !strings.Contains(detail, "1") {
		t.Errorf("the detail does not say how many are outstanding: %q", detail)
	}
}
