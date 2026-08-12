package obs_test

// a099 4라운드 리뷰가 낸 발송자 쪽 결함들.
//
// 공통 주제 하나: **임차를 놓는 경로와 잃는 경로가 서로 다른 말을 한다.** 어떤 곳은
// 결과를 버리고, 어떤 곳은 원장이 "이건 오류"라고 적은 값을 평범한 경합으로 강등하고,
// 종료 중에는 반납 자체가 실패한다.
//
// 초등학생 설명: 알림을 보내는 사람이 "내가 이 줄을 맡았다"는 이름표를 붙인다.
// 이 파일은 그 이름표를 **뗄 때** 생기는 문제들을 잡는다 — 못 떼거나, 떼면서
// 엉뚱한 말을 하거나, 남이 뗀 걸 내가 뗀 것처럼 구는 경우.

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	// 저널이 쓰는 것과 같은 드라이버. 한 테스트가 저널 파일에 두 번째 커넥션을 열어
	// 행을 지운다 — 원장에 테스트 전용 삭제 API 를 내지 않기 위해서다.
	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// logLines 는 JSON 로그 한 줄씩을 map 으로 돌려준다. 수준(level)까지 봐야 하는
// 테스트가 있어서 문자열 포함 검사로는 부족하다.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, raw := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("log line is not JSON: %q (%v)", raw, err)
		}
		out = append(out, m)
	}
	return out
}

// lineFor 는 주어진 event 이름의 첫 줄을 찾는다.
func lineFor(lines []map[string]any, event obs.EventType) map[string]any {
	for _, m := range lines {
		if m["event"] == string(event) {
			return m
		}
	}
	return nil
}

// cancelsOnFirstPublish 는 첫 발행에서 컨텍스트를 취소한다. 엔진이 발송 도중
// 내려가는 것 — 종료, 크래시 직전의 취소 — 을 테스트 안에서 결정적으로 만든다.
type cancelsOnFirstPublish struct {
	cancel context.CancelFunc
	calls  int
}

func (p *cancelsOnFirstPublish) Publish(context.Context, obs.Notification) error {
	p.calls++
	if p.calls == 1 && p.cancel != nil {
		p.cancel()
	}
	return errors.New("transport is down")
}

// TestACancelledSenderStillHandsTheLeaseBack.
//
// 4라운드 P1. deliver 의 주석은 "The lease does not stay"라고 단언하지만, 루프를
// 빠져나오는 두 길 중 하나가 **컨텍스트 취소**이고(:503 `if !n.wait(ctx) { break }`),
// 반납은 바로 그 죽은 컨텍스트로 발행된다. settleUnderClaim 은 그 ctx 로 트랜잭션을
// 열므로 커밋될 수가 없다.
//
// 결과: 배달 도중 내려간 엔진은 PENDING critical 행에 살아 있는 임차를 남긴다.
// 재기동한 프로세스의 첫 관측은 ClaimHeldElsewhere 로 떨어지고, 그 분기는 발행도
// 게이트도 에스컬레이션도 하지 않는다. 죽은 발송자가 알림을 삼킨다.
func TestACancelledSenderStillHandsTheLeaseBack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pub := &cancelsOnFirstPublish{cancel: cancel}
	n, j, _, _ := a099LoggedSender(t, pub, 3)

	_ = n.Notify(ctx, a099Event())

	// 읽기는 살아 있는 컨텍스트로 한다. 검사하려는 것은 행의 상태이지 취소가
	// 읽기까지 막는다는 사실이 아니다.
	pending, err := j.PendingAlerts(context.Background(), 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending rows = %d, want 1 — the alert must be preserved", len(pending))
	}
	if pending[0].ClaimedBy != "" || pending[0].ClaimExpiresAt != nil {
		t.Errorf("a sender cancelled mid-delivery kept the lease (by=%q, expires=%v). "+
			"The restarted process will find this row held by a sender that no longer "+
			"exists, and the held branch neither publishes nor escalates",
			pending[0].ClaimedBy, pending[0].ClaimExpiresAt)
	}
}

// TestAHeldRowIsNotWhispered.
//
// 4라운드 P0/P1. ClaimHeldElsewhere 는 "배제가 제대로 도는 중"이라는 전제로 info
// 한 줄만 남긴다. 그 전제는 발송자가 둘일 때만 맞다 — 오늘 배선에는 프로덕션
// Notifier 가 하나뿐이고 Notifier.Flush 는 프로덕션 호출자가 없다. 그래서 이
// 분기에 닿는 **유일한 실제 원인은 죽은 발송자가 남긴 임차**이고, 그 경우 조용한
// 것이 정확히 틀린 답이다.
//
// 회수 자체는 a099 의 몫이 아니다(a098 이 배달 루프를 만든다). a099 가 지는 책임은
// 그 상태를 만들지 않는 것과, 만들어졌을 때 **보이게** 하는 것이다.
func TestAHeldRowIsNotWhispered(t *testing.T) {
	pub := &alwaysFails{}
	n, j, buf, _ := a099LoggedSender(t, pub, 1)
	ctx := context.Background()

	// 다른(이제는 없는) 발송자가 그 행을 쥐고 있는 상태를 만든다.
	id, err := j.EnqueueAlert(ctx, journal.Alert{
		EventKey: "exit.proposal_refused|pos-1|LADDER_PARTIAL|2",
		Type:     "exit.proposal_refused",
		Severity: "critical",
	})
	if err != nil {
		t.Fatalf("arranging the row: %v", err)
	}
	if claim, err := j.ClaimAlertByID(ctx, id, "notifier:the-process-that-died"); err != nil {
		t.Fatalf("arranging the dead holder: %v", err)
	} else if claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("arranging the dead holder: disposition = %v", claim.Disposition)
	}

	if err := n.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if pub.calls != 0 {
		t.Fatalf("publishes = %d, want 0 — the row is held, so nothing may be sent", pub.calls)
	}

	line := lineFor(logLines(t, buf), obs.EventAlertClaimHeld)
	if line == nil {
		t.Fatalf("no engine.alert_claim_held line at all. The log was:\n%s", buf.String())
	}
	if line["level"] != "WARN" {
		t.Errorf("claim-held level = %v, want WARN — in this wiring the only way to reach "+
			"this branch is a lease left by a sender that died, and an unsent critical "+
			"alert is not an INFO", line["level"])
	}
	if _, ok := line["claim_expires_at"]; !ok {
		t.Error("the claim-held line does not say when the lease frees. That timestamp is " +
			"the operator's only estimate of when this alert can move")
	}
}

// TestALeaseLineCarriesItsOwnName.
//
// 4라운드가 **테스트를 쓰다가** 찾은 것이고, 이 라운드에서 가장 조용한 결함이다.
//
// Logger.emit 은 이미 `event` 에 **줄 자신의 이름**을 쓴다(log.go). 그런데 호출부가
// `FieldEvent, string(e.Type)` 로 **한 번 더** 그 키를 넘겼다. JSON 은 같은 키가 두 번
// 나오면 뒤엣것이 이긴다 — 그래서 engine.alert_claim_held 줄은 자기 이름을
// `exit.proposal_refused` 로 보고했다.
//
// 결과: a099 가 더한 이벤트 셋(held · lost · stolen)은 **매칭이 불가능했다.** 운영자가
// `event == "engine.alert_claim_held"` 로 규칙을 걸면 영원히 안 맞는다. event.go 가
// 이 셋을 따로 만든 이유가 *"a line that means three things cannot be alerted on"*
// 이었는데, 그 줄들은 자기 이름조차 달고 있지 않았다.
//
// 촉발한 조건은 `trigger_event` 로 옮겼다. 정보는 그대로 있고, `event` 는 log.go 의
// 문서가 처음부터 약속한 것 — 줄의 이름 — 이 된다.
func TestALeaseLineCarriesItsOwnName(t *testing.T) {
	pub := &alwaysFails{}
	n, j, buf, _ := a099LoggedSender(t, pub, 1)
	ctx := context.Background()

	id, err := j.EnqueueAlert(ctx, journal.Alert{
		EventKey: "exit.proposal_refused|pos-1|LADDER_PARTIAL|2",
		Type:     "exit.proposal_refused",
		Severity: "critical",
	})
	if err != nil {
		t.Fatalf("arranging the row: %v", err)
	}
	if _, err := j.ClaimAlertByID(ctx, id, "notifier:somebody-else"); err != nil {
		t.Fatalf("arranging the holder: %v", err)
	}
	if err := n.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	held := lineFor(logLines(t, buf), obs.EventAlertClaimHeld)
	if held == nil {
		t.Fatalf("the claim-held line does not report itself as %s, so no alerting rule "+
			"keyed on that name can ever match it. The log was:\n%s",
			obs.EventAlertClaimHeld, buf.String())
	}
	if held["trigger_event"] != string(obs.EventExitProposalRefused) {
		t.Errorf("trigger_event = %v, want %s — the condition that caused the line has to "+
			"survive the move off the `event` key", held["trigger_event"],
			obs.EventExitProposalRefused)
	}
}

// deletesTheRow 는 발행에 실패하면서, 그 사이에 원장에서 행을 지운다. 원장이
// 발송자가 쥐고 있던 행을 잃는 경우다.
//
// 지우는 쪽은 저널의 API 가 아니라 같은 파일에 연 두 번째 커넥션이다. 원장에
// 테스트 전용 삭제 구멍을 내는 것보다 낫다 — 그 구멍은 프로덕션에도 남는다.
type deletesTheRow struct {
	t    *testing.T
	path string
	once bool
}

func (p *deletesTheRow) Publish(context.Context, obs.Notification) error {
	if !p.once {
		p.once = true
		db, err := sql.Open("sqlite", p.path)
		if err != nil {
			p.t.Fatalf("arranging the disappearance: %v", err)
		}
		defer db.Close()
		if _, err := db.Exec(`DELETE FROM alert_outbox`); err != nil {
			p.t.Fatalf("arranging the disappearance: %v", err)
		}
	}
	return errors.New("transport is down")
}

// TestARowThatVanishedIsNotReportedAsContention.
//
// 4라운드. SettleNotFound 는 원장이 주석에 "That one *is* an error: the ledger lost
// a row a sender was holding"이라고 못박은 **유일한** 값이다. 그런데 실패-기록
// 경로는 그것을 `!= SettleApplied` 한 덩어리로 삼켜 logLeaseLost 로 보내고,
// logLeaseLost 에는 SettleNotFound 분기가 없다. 결과는 claimed_by 가 빈
// engine.alert_claim_lost 경고 — 즉 "평범한 경합"으로 보고된다.
//
// 같은 값이 성공 경로에서는 게이트를 잠그는 오류다. 한 사실에 두 결론은 둘 중
// 하나가 틀렸다는 뜻이다.
func TestARowThatVanishedIsNotReportedAsContention(t *testing.T) {
	n, j, buf, _ := a099LoggedSender(t, nil, 2)
	n.Publisher = &deletesTheRow{t: t, path: j.Path()}
	ctx := context.Background()

	_ = n.Notify(ctx, a099Event())

	lines := logLines(t, buf)
	if lost := lineFor(lines, obs.EventAlertClaimLost); lost != nil && lost["claimed_by"] == "" {
		t.Error("a row the ledger lost was reported as engine.alert_claim_lost naming no " +
			"holder — that is the shape of routine contention, and this is the one " +
			"outcome the ledger calls an error")
	}
	if lineFor(lines, obs.EventAlertUndelivered) == nil {
		t.Errorf("the ledger lost a row a sender was holding and no engine.alert_undelivered "+
			"line was written. The log was:\n%s", buf.String())
	}
}

// TestFlushHandsTheLeaseBackWhenThePublishFails.
//
// 4라운드. Flush 의 실패 경로는 정산 결과 둘과 반납 결과를 전부 `_, _ =` 로 버린다.
// deliver 는 같은 사실에 logLeaseLost 를 쓴다. 한 사실이 한 경로에서는 운영자에게
// 보이고 다른 경로에서는 안 보인다.
//
// 그리고 반납이 실제로 되는지도 아무 테스트가 안 본다. 반납이 빠지면 flush 가
// 실패한 행마다 임차가 만료까지 남고, 다음 flush 는 그 행을 "내 것이 아니다"라며
// 건너뛴다 — 배달되지 않은 critical 알림이, 그것을 건너뛰는 루프 때문에 안 보인다.
func TestFlushHandsTheLeaseBackWhenThePublishFails(t *testing.T) {
	pub := &alwaysFails{}
	n, j, _, _ := a099LoggedSender(t, pub, 1)
	ctx := context.Background()

	if _, err := j.EnqueueAlert(ctx, journal.Alert{
		EventKey: "exit.proposal_refused|pos-9|LADDER_PARTIAL|1",
		Type:     "exit.proposal_refused",
		Severity: "critical",
	}); err != nil {
		t.Fatalf("arranging the row: %v", err)
	}

	delivered, _, err := n.Flush(ctx)
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if delivered != 0 {
		t.Fatalf("delivered = %d, want 0 — the transport is down", delivered)
	}

	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending rows = %d, want 1", len(pending))
	}
	if pending[0].ClaimedBy != "" {
		t.Fatalf("a failed flush kept the lease (%q) — the next cycle skips this row",
			pending[0].ClaimedBy)
	}

	// 그리고 "안 잠겼다"가 말대로인지: 전송로가 돌아오면 다음 flush 가 보낸다.
	ok := &deliversOnce{}
	n.Publisher = ok
	delivered, _, err = n.Flush(ctx)
	if err != nil {
		t.Fatalf("second Flush: %v", err)
	}
	if delivered != 1 {
		t.Errorf("second flush delivered = %d, want 1 — the row the first flush gave up on "+
			"has to be sendable once the transport is back", delivered)
	}
}

// deliversOnce 는 정상 전송로다.
type deliversOnce struct{ calls int }

func (p *deliversOnce) Publish(context.Context, obs.Notification) error {
	p.calls++
	return nil
}

// TestALeaseShorterThanTheBudgetIsReported.
//
// 4라운드. lease > 배달 예산은 이 설계의 안전 성질인데, 그것을 지키는 것은
// **기본값에서만** 도는 테스트 하나다. Notifier.Attempts·RetryDelay·Ntfy.Timeout 은
// 전부 exported 이고 journal.Options.AlertLease 도 그렇다. 둘을 비교하는 코드는
// 어디에도 없다.
//
// 예산이 임차보다 길면, 발송자는 이 코드가 쓰라고 시킨 재시도 예산을 쓰는 도중에
// 자기 행을 잃는다. 그러면 두 발송자가 발행하는데 **아무 데도 잘못된 것이 없다** —
// 그게 이 change 가 없애려던 바로 그 상태다.
func TestALeaseShorterThanTheBudgetIsReported(t *testing.T) {
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	buf := &bytes.Buffer{}
	n := &obs.Notifier{
		Log:       obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clk}),
		Publisher: &deliversOnce{},
		Journal:   j,
		Gate:      execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
		Clock:     clock.System(),
		// 기본 임차(81s)를 확실히 넘기는 예산.
		Attempts:    40,
		RetryDelay:  time.Minute,
		RemindAfter: a099Remind,
	}

	if err := n.Notify(context.Background(), a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if lineFor(logLines(t, buf), obs.EventAlertLeaseTooShort) == nil {
		t.Errorf("a notifier whose retry budget (%v) outlasts the lease it is issued (%v) "+
			"reported nothing. It will lose its row mid-budget and a second sender will "+
			"publish, with no misconfiguration recorded anywhere",
			obs.AlertDeliveryBound(40, obs.DefaultPublishTimeout, time.Minute, journal.DefaultBusyTimeout),
			journal.DefaultAlertLease)
	}
}
