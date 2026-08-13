package engine

// a098 4.4 (엔진 쪽) — 운영자가 밀린 것을 읽고, 승인으로 게이트를 푼다.
//
// 여기서 세 R 이 만난다. 셋을 한 파일에 두는 이유는 **서로를 반증하기 때문**이다:
//
//	R9   목록이 임차 보유자와 나이를 보여 주고 **본문·payload·토큰·계좌는 안 보여 준다**
//	R5   승인이 게이트를 풀고 **운영자 이름이 원장에 실제로 들어간다**
//	R17  **전부 승인해도** 「발송 주체 없음」 차단은 **안 풀린다**
//
// R5 만 있으면 `Gate.Clear`를 무조건 부르는 구현이 통과하고, 그것은 R17 을 깬다.
// R9 만 있으면 아무것도 안 돌려주는 목록이 통과한다 — 그래서 **있어야 할 것**과
// **없어야 할 것**을 같은 테스트에서 본다.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a098Ops builds the operator surface on the **production** wiring.
//
// buildGateway 가 만든 notifier·gate 를 그대로 쓴다. 손으로 조립한 notifier 로
// 재면 「게이트를 푸는 것은 notifier 다」라는 성질이 테스트 안에서만 참일 수 있다.
func a098Ops(t *testing.T, j *journal.Journal) (*AlertOperations, *execgw.EntryGate) {
	t.Helper()
	// buildGateway 는 투영 훅이 안 묶여 있으면 조립을 거절한다(ErrProjectionUnbound) —
	// 그 거절은 이 change 와 무관한 기존 계약이고, R8 의 테스트도 같은 줄을 먼저 부른다.
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	wiring := a098BuildGateway(t, context.Background(), j)
	return &AlertOperations{
		journal:  j,
		notifier: wiring.notifier,
		gate:     wiring.entry,
		clock:    clock.System(),
	}, wiring.entry
}

// a098AlertContentSentinels are strings no operator listing may repeat. 각각
// journal.Alert 의 서로 다른 칸에 심는다 — 하나만 심으면 「그 칸만 뺀 구현」이 통과한다.
var a098AlertContentSentinels = map[string]string{
	"title":   "SENTINEL-TITLE-005930",
	"body":    "SENTINEL-BODY-account-11122233",
	"payload": "SENTINEL-PAYLOAD-{\"qty\":7}",
	"key":     "SENTINEL-KEY-a098",
}

func a098SeedContentfulAlert(t *testing.T, j *journal.Journal) int64 {
	t.Helper()
	id, err := j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: a098AlertContentSentinels["key"],
		Type:     "engine.exit_proposal_refused",
		Severity: "critical",
		Title:    a098AlertContentSentinels["title"],
		Body:     a098AlertContentSentinels["body"],
		Payload:  a098AlertContentSentinels["payload"],
	})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

// TestThePendingListingCarriesNoAlertContent is R9's forbidding half.
//
// **직렬화한 결과를 본다.** 구조체 필드만 세면 `journal.Alert` 를 그대로 담은 필드
// 하나를 놓치고, 그 하나가 CLI 화면에 전부를 싣는다 — 6라운드 P4 가 지적한 형태다.
// 원천 질의가 `body` 와 `payload` 를 **이미 싣고 온다**(outbox.go alertSelect)는 것이
// 이 검사가 필요한 이유다: 안 실어 오는 것이 아니라 **안 내보내는 것**이 성질이다.
func TestThePendingListingCarriesNoAlertContent(t *testing.T) {
	j := openTestJournal(t)
	a098SeedContentfulAlert(t, j)
	ops, _ := a098Ops(t, j)

	rows, err := ops.Pending(context.Background())
	if err != nil {
		t.Fatalf("Pending: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Pending returned %d rows, want 1", len(rows))
	}

	encoded, err := json.Marshal(rows)
	if err != nil {
		t.Fatalf("marshalling the listing: %v", err)
	}
	for column, sentinel := range a098AlertContentSentinels {
		if strings.Contains(string(encoded), sentinel) {
			t.Errorf("the listing leaks the alert's %s: %s", column, encoded)
		}
	}
	// 그리고 **칸 이름**을 따로 본다. 값만 보면 내용이 비어 있는 행에서 통과하고,
	// 문자열 포함으로 보면 값이 우연히 "body" 를 담아도 걸린다 — 둘 다 측정이 아니다.
	// 그래서 키를 열거한다.
	var decoded []map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decoding the listing: %v", err)
	}
	allowed := map[string]bool{
		"id": true, "type": true, "age_seconds": true, "attempts": true,
		"claimed_by": true, "claim_age_seconds": true, "claim_expired": true,
	}
	for _, row := range decoded {
		for key := range row {
			if !allowed[key] {
				t.Errorf("the listing has an unexpected field %q — 투영이 넓어졌다: %s", key, encoded)
			}
		}
	}
}

// TestThePendingListingShowsTheLeaseHolderAndItsAge is R9's requiring half.
//
// 없어야 할 것만 재면 **빈 목록이 통과한다.** 네 칸이 실제로 차는 것을 함께 본다.
func TestThePendingListingShowsTheLeaseHolderAndItsAge(t *testing.T) {
	j := openTestJournal(t)
	id := a098SeedContentfulAlert(t, j)
	ctx := context.Background()
	claim, err := j.ClaimAlertByID(ctx, id, "engine.alert_delivery")
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	// ⛔ 토큰은 **이 반환에서만** 얻는다. 같은 청구자로 다시 claim 하면
	// `HeldElsewhere` 로 물러난다 — a099 의 `alertClaimable` 은 청구자 자신을
	// 특별대우하지 않는다(R1 실측). 다시 부르는 우회는 빈 토큰을 준다.
	if _, err := j.MarkAlertAttemptFailed(ctx, id, claim.Token, "transport refused"); err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}
	// 시도 실패는 임차를 안 돌려준다 — 보유자 칸이 차 있어야 R9 를 잴 수 있다.
	if _, err := j.ClaimAlertByID(ctx, id, "engine.alert_delivery"); err != nil {
		t.Fatalf("re-claiming after a failed attempt: %v", err)
	}

	ops, _ := a098Ops(t, j)
	rows, err := ops.Pending(ctx)
	if err != nil {
		t.Fatalf("Pending: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Pending returned %d rows, want 1", len(rows))
	}
	row := rows[0]
	if row.ID != id {
		t.Errorf("id = %d, want %d", row.ID, id)
	}
	if row.Type != "engine.exit_proposal_refused" {
		t.Errorf("type = %q — 어느 조건이 안 나갔는지 말하지 못한다", row.Type)
	}
	if row.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", row.Attempts)
	}
	if row.ClaimedBy != "engine.alert_delivery" {
		t.Errorf("claimed_by = %q, want the sender holding the lease — "+
			"보유자가 없으면 「막힌 것」과 「보내는 중」이 같아 보인다", row.ClaimedBy)
	}
	if row.ClaimExpired {
		t.Error("claim_expired is true on a lease just taken")
	}
	if row.AgeSeconds < 0 {
		t.Errorf("age_seconds = %d", row.AgeSeconds)
	}
}

// TestAcknowledgingClearsTheBacklogLatchAndNamesTheOperator is R5's engine half.
//
// **이름이 원장에 실제로 들어간 것까지 본다.** 게이트만 보면 `AcknowledgeAlert` 를
// 안 부르고 `Clear` 만 하는 구현이 통과하고, 그러면 승인 기록이 안 남는다 —
// 원장은 빈 이름을 이미 거부하므로(outbox.go:482) 위험은 빈 actor 가 아니라
// **audit trail 이 조용히 비는 것**이다(6라운드 P4).
func TestAcknowledgingClearsTheBacklogLatchAndNamesTheOperator(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	first := a098SeedContentfulAlert(t, j)
	second := a098ParkedAlert(t, j, "a098-ack-second")

	ops, gate := a098Ops(t, j)
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")

	result, err := ops.Acknowledge(ctx, "kim.operator")
	if err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if result.RemainingUndelivered != 0 {
		t.Errorf("remaining = %d, want 0", result.RemainingUndelivered)
	}
	if len(result.EntryBlocks) != 0 {
		t.Errorf("entry is still blocked by %v after the backlog was acknowledged", result.EntryBlocks)
	}
	if blocks := gate.Blocks(); len(blocks) != 0 {
		t.Errorf("gate.Blocks() = %v, want empty — 승인이 진입을 안 풀었다", blocks)
	}
	for _, id := range []int64{first, second} {
		row, err := j.LookupAlert(ctx, id)
		if err != nil {
			t.Fatalf("LookupAlert(%d): %v", id, err)
		}
		if row.AcknowledgedBy != "kim.operator" {
			t.Errorf("alert %d acknowledged_by = %q, want the operator's name — "+
				"게이트만 풀고 원장에 안 적으면 audit trail 이 죽는다", id, row.AcknowledgedBy)
		}
		if row.AcknowledgedAt == nil {
			t.Errorf("alert %d has no acknowledged_at", id)
		}
	}
}

// TestAcknowledgingEverythingLeavesTheSenderDownLatch is R17.
//
// 오늘 사유가 하나뿐이라 `Acknowledge` 가 미전달 0 에서 **그 하나를 푼다**
// (notifier.go:874-876). 4.3.3 이 더한 두 번째 사유는 그 함수가 이름으로 모른다 —
// **그것이 성질이고, 이 테스트가 그것을 고정한다.** 안 그러면 「전부 승인」이
// 아무도 안 보내는 엔진의 진입을 열어 준다(사용자 결정 8-1).
//
// ⛔ 발송 주체 없음 래치를 **손으로 걸지 않는다.** 등록부가 요구하는 관측은
// *"실행자를 죽이고 밀린 행을 전부 승인한 뒤"*이고, 손으로 건 래치는 4.3 의 경로가
// 실제로 그 사유를 거는지를 안 본다 — 그러면 R17 은 `Clear` 만 재는 테스트가 된다.
func TestAcknowledgingEverythingLeavesTheSenderDownLatch(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	a098SeedContentfulAlert(t, j)

	ops, gate := a098Ops(t, j)
	// 밀린 것 쪽 래치는 notifier 의 경로이고 이 테스트가 재는 것이 아니다 — 직접 건다.
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")

	// 발송 주체 없음 쪽은 **실행자를 실제로 죽여서** 얻는다.
	returned := make(chan struct{})
	a098RunWithAux(t, a098DeliveryAux(gate, func(context.Context) error {
		close(returned)
		return errors.New("the delivery executor stopped")
	}), a098LiveLoop("exit"), &a098CountingEscalation{})
	<-returned
	if !a098WaitForBlock(gate, execgw.ReasonAlertSenderDown, 2*time.Second) {
		t.Fatal("the executor's death did not latch critical_alert_sender_down; " +
			"R17 이 잴 상태가 안 만들어졌다")
	}

	result, err := ops.Acknowledge(ctx, "kim.operator")
	if err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if result.RemainingUndelivered != 0 {
		t.Fatalf("remaining = %d, want 0 — 승인이 안 됐으면 R17 이 공허하게 통과한다",
			result.RemainingUndelivered)
	}
	blocks := gate.Blocks()
	if _, still := blocks[execgw.ReasonAlertSenderDown]; !still {
		t.Fatal("acknowledging the backlog cleared critical_alert_sender_down; " +
			"밀린 것이 없어졌을 뿐 보내는 주체는 여전히 없다")
	}
	if _, gone := blocks[execgw.ReasonAlertUndelivered]; gone {
		t.Error("the backlog latch survived the acknowledgement; 대조군이 대조군이 아니다")
	}
	want := string(execgw.ReasonAlertSenderDown)
	if len(result.EntryBlocks) != 1 || result.EntryBlocks[0] != want {
		t.Errorf("EntryBlocks = %v, want [%s] — 운영자에게 「아직 막혀 있다」를 말해야 한다",
			result.EntryBlocks, want)
	}
}

// TestAcknowledgingRefusesABlankOperator pins the audit trail's only guard.
func TestAcknowledgingRefusesABlankOperator(t *testing.T) {
	j := openTestJournal(t)
	a098SeedContentfulAlert(t, j)
	ops, gate := a098Ops(t, j)
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")

	if _, err := ops.Acknowledge(context.Background(), "   "); err == nil {
		t.Fatal("a blank operator was accepted; 누가 봤는지 없는 승인은 승인이 아니다")
	}
	if _, still := gate.Blocks()[execgw.ReasonAlertUndelivered]; !still {
		t.Error("the refused acknowledgement cleared the latch anyway")
	}
}

// TestTheOperatorSurfaceRefusesAnIncompleteEngine covers the constructor.
//
// 부분 조립을 받아들이면 승인이 원장에만 남고 진입은 막힌 채가 된다 — D7.1 이
// 없애려는 바로 그 상태가 nil 검사를 통해 돌아온다.
func TestTheOperatorSurfaceRefusesAnIncompleteEngine(t *testing.T) {
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	wiring := a098BuildGateway(t, context.Background(), j)

	cases := map[string]*Context{
		"nil context":  nil,
		"no journal":   {Notifier: wiring.notifier, Entry: wiring.entry},
		"no notifier":  {Journal: j, Entry: wiring.entry},
		"no gate":      {Journal: j, Notifier: wiring.notifier},
		"nothing wire": {},
	}
	for name, ctx := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ctx.AlertOperations(); err == nil {
				t.Fatal("AlertOperations accepted an incomplete engine")
			}
		})
	}

	full := &Context{Journal: j, Notifier: wiring.notifier, Entry: wiring.entry}
	ops, err := full.AlertOperations()
	if err != nil {
		t.Fatalf("AlertOperations on a complete engine: %v", err)
	}
	if ops.clock == nil {
		t.Error("the surface has no clock; 나이 칸이 계산될 수 없다")
	}
}

// TestAnAgeIsNeverNegative pins the clamp.
func TestAnAgeIsNeverNegative(t *testing.T) {
	now := time.Now()
	if got := elapsedSeconds(now.Add(4*time.Second), now); got != 0 {
		t.Errorf("elapsedSeconds(future) = %d, want 0", got)
	}
	if got := elapsedSeconds(now.Add(-90*time.Second), now); got != 90 {
		t.Errorf("elapsedSeconds(-90s) = %d, want 90", got)
	}
}
