package engine

// a124 — 계속 실패하는 배달 실행자는 진입을 막는다 (tasks 2.1 · 2.2 · 2.3 · 2.4 · 2.5 · 2.8 · 2.9 · 2.10 · 2.11 · 2.12).
//
// HEAD 의 실행자는 행을 몇 번 실패해도 게이트도 모드도 건드리지 않는다(`alertdelivery.go` 에 Gate ·
// Escalate 0 회). 동기 경로 `Notifier.deliver` 가 그 일을 하는데, a092 가 그 시도를 엔진 루프에서
// 빼면 주인이 사라진다. 이 파일은 실행자가 그 주인이 되었는지를 **동기 경로가 한 번도 시도하지 않는
// 구성**에서 잰다 — 어떤 시험도 Notifier 의 발송을 거치지 않는다.
//
// 결정적 순서가 필요한 시험(원칙 E)은 실행자의 판정 단계 훅(`judgeHook`)으로 해제를 끼워 넣는다.
// 원장 결함은 원장 래퍼(`ledger`)로 주입한다 — 원장 자체의 결함 주입은 journal 패키지 시험(2.7)이 잰다.

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const a124Account = "acct-a124"

// a124Ledger 는 실제 원장을 감싸고 원하는 호출만 실패시키거나 결과를 바꾼다.
type a124Ledger struct {
	*journal.Journal

	mu             sync.Mutex
	failList       int // 남은 나열 실패 수 (-1 = 끝없이)
	failClaim      int
	failRecord     int
	failRecordFor  map[int64]bool // 이 행들의 실패 기록만 오류
	failDeliver    int
	recordOutcome  *journal.SettleOutcome // 설정되면 실제 기록 없이 이 결과
	deliverOutcome *journal.SettleOutcome
	failEscalate   bool
	escalations    int
	faultErr       error
	beforeClaim    func(id int64) // 나열 뒤 · 임차 전에 끼는 일(다른 발송자 · 승인)
	beforeRecord   func(id int64) // 발행 뒤 · 실패 기록 전에 끼는 일
	beforeDeliver  func(id int64) // 발행 뒤 · 전달 정산 전에 끼는 일
}

func (l *a124Ledger) take(n *int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	switch {
	case *n < 0:
		return true
	case *n > 0:
		*n--
		return true
	}
	return false
}

func (l *a124Ledger) fault() error {
	if l.faultErr != nil {
		return l.faultErr
	}
	return errors.New("a124 synthetic ledger fault")
}

func (l *a124Ledger) PendingAlertsForDelivery(ctx context.Context, limit, attemptLimit int) ([]journal.Alert, error) {
	if l.take(&l.failList) {
		return nil, l.fault()
	}
	return l.Journal.PendingAlertsForDelivery(ctx, limit, attemptLimit)
}

func (l *a124Ledger) ClaimAlertByID(ctx context.Context, id int64, claimant string) (journal.ClaimResult, error) {
	if l.beforeClaim != nil {
		hook := l.beforeClaim
		l.beforeClaim = nil
		hook(id)
	}
	if l.take(&l.failClaim) {
		return journal.ClaimResult{ID: id}, l.fault()
	}
	return l.Journal.ClaimAlertByID(ctx, id, claimant)
}

func (l *a124Ledger) MarkAlertAttemptFailed(ctx context.Context, id int64, token, cause string) (journal.SettleResult, error) {
	if l.beforeRecord != nil {
		hook := l.beforeRecord
		l.beforeRecord = nil
		hook(id)
	}
	if l.take(&l.failRecord) || l.failRecordFor[id] {
		return journal.SettleResult{}, l.fault()
	}
	if l.recordOutcome != nil {
		return journal.SettleResult{Outcome: *l.recordOutcome}, nil
	}
	return l.Journal.MarkAlertAttemptFailed(ctx, id, token, cause)
}

func (l *a124Ledger) MarkAlertDelivered(ctx context.Context, id int64, token string) (journal.SettleResult, error) {
	if l.beforeDeliver != nil {
		hook := l.beforeDeliver
		l.beforeDeliver = nil
		hook(id)
	}
	if l.take(&l.failDeliver) {
		return journal.SettleResult{}, l.fault()
	}
	if l.deliverOutcome != nil {
		return journal.SettleResult{Outcome: *l.deliverOutcome}, nil
	}
	return l.Journal.MarkAlertDelivered(ctx, id, token)
}

func (l *a124Ledger) EscalateOperatingMode(ctx context.Context, accountRef, trigger string,
	announcer journal.ModeAnnouncer) (journal.OperatingModeRecord, bool, error) {
	l.mu.Lock()
	l.escalations++
	fail := l.failEscalate
	l.mu.Unlock()
	if fail {
		return journal.OperatingModeRecord{}, false, l.fault()
	}
	return l.Journal.EscalateOperatingMode(ctx, accountRef, trigger, announcer)
}

func (l *a124Ledger) escalationCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.escalations
}

type a124Fixture struct {
	t    *testing.T
	path string
	j    *journal.Journal
	clk  *clock.Fake
	gate *execgw.EntryGate
	led  *a124Ledger
	d    *alertDeliverer
	logs *bytes.Buffer
}

func a124Setup(t *testing.T, pub obs.Publisher) *a124Fixture {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), journal.DBFileName)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	led := &a124Ledger{Journal: j}
	logs := &bytes.Buffer{}
	d := &alertDeliverer{
		Journal: j, Publisher: pub, Clock: clk,
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a124-test",
		Log:  obs.NewLogger(obs.LogOptions{Writer: logs, JSON: true, Clock: clk}),
		Gate: gate, AccountRef: a124Account, ledger: led,
	}
	return &a124Fixture{t: t, path: path, j: j, clk: clk, gate: gate, led: led, d: d, logs: logs}
}

func (f *a124Fixture) row(t *testing.T, key string) int64 {
	t.Helper()
	id, err := f.j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: key, Type: "execgw.order_unresolved", Severity: "critical",
		Title: "UNRESOLVED_IN_DOUBT", Body: "an attempt is unresolved",
	})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

// exhaust 는 행의 attempts 를 n 으로 올리고 임차를 돌려준다(실행자 밖의 발송자가 실패한 것처럼).
func (f *a124Fixture) exhaust(t *testing.T, id int64, n int) {
	t.Helper()
	ctx := context.Background()
	claim, err := f.j.ClaimAlertByID(ctx, id, "a124-other")
	if err != nil || claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("claim for exhaust: %+v %v", claim, err)
	}
	for i := 0; i < n; i++ {
		if _, err := f.j.MarkAlertAttemptFailed(ctx, id, claim.Token, "earlier failure"); err != nil {
			t.Fatalf("exhaust attempt: %v", err)
		}
	}
	if _, err := f.j.ReleaseAlertClaim(ctx, id, claim.Token); err != nil {
		t.Fatalf("exhaust release: %v", err)
	}
}

func (f *a124Fixture) cycles(t *testing.T, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		_ = f.d.cycle(context.Background())
	}
}

func (f *a124Fixture) latched() bool {
	_, ok := f.gate.Blocks()[execgw.ReasonAlertUndelivered]
	return ok
}

func (f *a124Fixture) mode(t *testing.T) string {
	t.Helper()
	snap, err := f.j.CurrentOperatingMode(context.Background(), a124Account)
	if err != nil {
		t.Fatalf("CurrentOperatingMode: %v", err)
	}
	return snap.Mode
}

func (f *a124Fixture) attempts(t *testing.T, id int64) int {
	t.Helper()
	row, err := f.j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	return row.Attempts
}

func a124Failing() *a098RecordingPublisher {
	return &a098RecordingPublisher{fail: errors.New("transport is down")}
}

// --- 2.1 ---------------------------------------------------------------------------------------

func TestTheExecutorLatchesAndEscalatesAtTheAttemptLimit(t *testing.T) {
	for _, tc := range []struct {
		name string
		pub  obs.Publisher
	}{{"transport fails", a124Failing()}, {"no publisher", nil}} {
		t.Run(tc.name, func(t *testing.T) {
			f := a124Setup(t, tc.pub)
			id := f.row(t, "a124-2.1")
			f.cycles(t, alertAttemptLimit-1)
			if f.latched() || f.mode(t) != journal.ModeNormal {
				t.Fatalf("latched=%v mode=%s below the limit (attempts=%d)", f.latched(), f.mode(t), f.attempts(t, id))
			}
			f.cycles(t, 1)
			if got := f.attempts(t, id); got != alertAttemptLimit {
				t.Fatalf("attempts = %d, want %d — every cycle is one attempt, a missing publisher included (D3)", got, alertAttemptLimit)
			}
			if !f.latched() {
				t.Fatal("the gate is open after the attempt limit — nobody latched it without the synchronous path")
			}
			if got := f.mode(t); got != journal.ModeEntryBlocked {
				t.Fatalf("mode = %s, want %s", got, journal.ModeEntryBlocked)
			}
		})
	}
}

// --- 2.2 · 2.3 ---------------------------------------------------------------------------------

func TestAnAcknowledgementClearsTheLatchAndTheModeRowStays(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.2")
	f.cycles(t, alertAttemptLimit)
	if !f.latched() {
		t.Fatal("arrangement: not latched")
	}
	n := &obs.Notifier{Journal: f.j, Gate: f.gate}
	if err := n.Acknowledge(context.Background(), "operator"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if f.latched() {
		t.Fatal("the latch survived an acknowledgement that emptied the backlog")
	}
	if got := f.mode(t); got != journal.ModeEntryBlocked {
		t.Fatalf("mode row = %s after acknowledgement, want %s (acknowledgement never relaxes a mode)", got, journal.ModeEntryBlocked)
	}
}

func TestARestartRelatchesFromThePendingRow(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.3")
	f.cycles(t, alertAttemptLimit)
	fresh := execgw.NewEntryGate(f.clk, map[execgw.RequiredQuery]time.Duration{})
	if err := restoreAlertEntryLatch(context.Background(), f.j, fresh); err != nil {
		t.Fatalf("restoreAlertEntryLatch: %v", err)
	}
	if _, ok := fresh.Blocks()[execgw.ReasonAlertUndelivered]; !ok {
		t.Fatal("the restarted gate is open with the exhausted row still PENDING")
	}
	if got := f.mode(t); got != journal.ModeEntryBlocked {
		t.Fatalf("mode row = %s after restart, want it still in the ledger", got)
	}
}

// --- 2.4 · 2.5 ---------------------------------------------------------------------------------

type a124TitlePublisher struct {
	mu   sync.Mutex
	sent []string
}

func (p *a124TitlePublisher) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = append(p.sent, n.Title)
	return errors.New("still down")
}

func TestANewRowIsTriedBeforeExhaustedOnes(t *testing.T) {
	pub := &a124TitlePublisher{}
	f := a124Setup(t, pub)
	ctx := context.Background()
	for i := 0; i < alertDeliveryBatch; i++ {
		id := f.row(t, "a124-2.4-old-"+string(rune('a'+i)))
		f.exhaust(t, id, alertAttemptLimit)
	}
	if _, err := f.j.EnqueueAlert(ctx, journal.Alert{EventKey: "a124-2.4-new", Type: "x", Severity: "critical", Title: "NEW"}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	f.cycles(t, 1)
	pub.mu.Lock()
	first := ""
	if len(pub.sent) > 0 {
		first = pub.sent[0]
	}
	pub.mu.Unlock()
	if first != "NEW" {
		t.Fatalf("the first row tried this cycle was %q, want the new one — exhausted rows are starving it", first)
	}
	pending, err := f.j.UndeliveredCount(ctx)
	if err != nil || pending != alertDeliveryBatch+1 {
		t.Fatalf("undelivered = %d (%v), want every exhausted row still counted", pending, err)
	}
}

// --- 2.8 ---------------------------------------------------------------------------------------

type a124AckingPublisher struct {
	j  *journal.Journal
	id int64
}

func (p *a124AckingPublisher) Publish(ctx context.Context, _ obs.Notification) error {
	_ = p.j.AcknowledgeAlert(ctx, p.id, "operator")
	return errors.New("down, and the operator acknowledged meanwhile")
}

func TestAnAcknowledgementDuringTheSendDoesNotLatch(t *testing.T) {
	f := a124Setup(t, nil)
	id := f.row(t, "a124-2.8")
	f.exhaust(t, id, alertAttemptLimit-1)
	f.d.Publisher = &a124AckingPublisher{j: f.j, id: id}
	f.cycles(t, 1)
	if f.latched() || f.led.escalationCount() != 0 {
		t.Fatalf("latched=%v escalations=%d after an acknowledgement that beat the failure", f.latched(), f.led.escalationCount())
	}
}

func TestOutcomesThatWroteNothingNeverJudge(t *testing.T) {
	for _, out := range []journal.SettleOutcome{journal.SettleLeaseLost, journal.SettleAlreadySettled} {
		t.Run(out.String(), func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			id := f.row(t, "a124-2.8-"+out.String())
			f.exhaust(t, id, alertAttemptLimit+2)
			o := out
			f.led.recordOutcome = &o
			f.cycles(t, 2)
			if f.latched() || f.led.escalationCount() != 0 {
				t.Fatalf("%s latched=%v escalations=%d", out, f.latched(), f.led.escalationCount())
			}
		})
	}
}

func TestARecordThatFindsNoRowLatchesButNeverEscalates(t *testing.T) {
	for _, out := range []journal.SettleOutcome{journal.SettleNotFound, journal.SettleOutcome(97)} {
		t.Run(out.String(), func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			f.row(t, "a124-2.8-nf-"+out.String())
			o := out
			f.led.recordOutcome = &o
			f.cycles(t, 1)
			if !f.latched() {
				t.Fatalf("%s did not latch (fail-safe: an unaccountable record is a latch)", out)
			}
			if f.led.escalationCount() != 0 || f.mode(t) != journal.ModeNormal {
				t.Fatalf("%s escalated (escalations=%d mode=%s) — the synchronous path does not (N6)", out, f.led.escalationCount(), f.mode(t))
			}
		})
	}
}

// --- 2.9 원칙 E ----------------------------------------------------------------------------------

// a124AtStage 는 판정 단계 훅에서 한 번만 fn 을 부른다.
func a124AtStage(d *alertDeliverer, stage alertJudgeStage, fn func()) {
	var once sync.Once
	d.judgeHook = func(s alertJudgeStage, _ int64) {
		if s == stage {
			once.Do(fn)
		}
	}
}

func (f *a124Fixture) acknowledgeAll(t *testing.T) {
	t.Helper()
	if err := (&obs.Notifier{Journal: f.j, Gate: f.gate}).Acknowledge(context.Background(), "operator"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
}

// (a) 해제가 근거 **앞** (㉬): 옛 해제 뒤에 새 행이 한도에 이른다 → 차단 + 승격.
func TestAClearBeforeTheEvidenceDoesNotDropTheJudgement(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.9a")
	f.cycles(t, alertAttemptLimit-1)
	f.gate.Clear(execgw.ReasonAlertUndelivered) // 옛 승인의 해제 — 근거보다 앞
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — a clear that came before the evidence dropped it", f.latched(), f.mode(t))
	}
}

// (b) 해제가 **뒤** (㉫ · 「세대 읽기 → 해제 → 적용」) → 차단 없음 + 승격, 그리고 제때 적용 대조와 같은 상태.
func TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave(t *testing.T) {
	late := a124Setup(t, a124Failing())
	late.row(t, "a124-2.9b")
	late.cycles(t, alertAttemptLimit-1)
	a124AtStage(late.d, alertStageEpochRead, func() { late.acknowledgeAll(t) })
	late.cycles(t, 1)

	onTime := a124Setup(t, a124Failing())
	onTime.row(t, "a124-2.9b")
	onTime.cycles(t, alertAttemptLimit) // 제때: 차단 + 승격
	onTime.acknowledgeAll(t)            // 그 뒤 같은 해제

	if late.latched() != onTime.latched() || late.mode(t) != onTime.mode(t) {
		t.Fatalf("late: latched=%v mode=%s · on time: latched=%v mode=%s — principle E broken",
			late.latched(), late.mode(t), onTime.latched(), onTime.mode(t))
	}
	if late.latched() || late.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("late: latched=%v mode=%s, want no latch and the escalation kept (Y1)", late.latched(), late.mode(t))
	}
}

// ㉣ 래치가 서기 전 전체 승인 — 지울 래치가 없어도 해제 요청은 세대를 올린다 → 빈 backlog 재잠금 없음, 승격은 제때처럼 남음.
func TestAFullAcknowledgementBeforeTheLatchLeavesNoLatchOnAnEmptyBacklog(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.9d")
	f.cycles(t, alertAttemptLimit-1)
	a124AtStage(f.d, alertStageEpochRead, func() { f.acknowledgeAll(t) })
	f.cycles(t, 1)
	if n, _ := f.j.UndeliveredCount(context.Background()); n != 0 {
		t.Fatalf("arrangement: backlog = %d, want 0", n)
	}
	if f.latched() {
		t.Fatal("the gate latched over an empty backlog the operator had just cleared")
	}
	if f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s — the on-time escalation would have stayed", f.mode(t))
	}
}

// AB2 허용 예외: 「정산 → 해제 → 세대 읽기 → 적용」 은 해제를 「앞」 으로 본다 → 보수적 재잠금.
func TestAClearBetweenTheEvidenceAndTheEpochReadRelatchesConservatively(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.9-window")
	f.cycles(t, alertAttemptLimit-1)
	a124AtStage(f.d, alertStageSettled, func() { f.gate.Clear(execgw.ReasonAlertUndelivered) })
	f.cycles(t, 1)
	if !f.latched() {
		t.Fatal("a clear inside the evidence-to-epoch window dropped the latch — the permitted error runs the other way")
	}
}

// ㉢ 무관한 승인(다른 행만, 미전달 남음)은 해제 요청이 아니다 → 잠금.
func TestAnUnrelatedAcknowledgementDoesNotDropTheJudgement(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.9c")
	other := f.row(t, "a124-2.9c-other")
	f.exhaust(t, other, 0)
	f.cycles(t, alertAttemptLimit-1)
	a124AtStage(f.d, alertStageEpochRead, func() {
		if err := (&obs.Notifier{Journal: f.j, Gate: f.gate}).Acknowledge(context.Background(), "operator", other); err != nil {
			t.Errorf("Acknowledge(other): %v", err)
		}
	})
	f.cycles(t, 1)
	if !f.latched() {
		t.Fatal("an acknowledgement that left rows undelivered dropped the judgement")
	}
}

// 실패 기록 NotFound 는 「뒤」 해제에서도 승격하지 않는다.
func TestALatchOnlyJudgementNeverEscalatesEvenAfterAClear(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.9-nf")
	nf := journal.SettleNotFound
	f.led.recordOutcome = &nf
	a124AtStage(f.d, alertStageEpochRead, func() { f.gate.Clear(execgw.ReasonAlertUndelivered) })
	f.cycles(t, 1)
	if f.latched() || f.led.escalationCount() != 0 {
		t.Fatalf("latched=%v escalations=%d", f.latched(), f.led.escalationCount())
	}
}

// Z2 · AA1: 승격 쓰기 실패면 조건부 차단 결과와 무관하게 무조건 차단 — 「차단 성공 → 해제 → 승격 실패」 포함.
func TestAFailedEscalationLeavesTheLatch(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stage alertJudgeStage
	}{{"block rejected by a clear", alertStageEpochRead}, {"block applied then cleared", alertStageLatched}} {
		t.Run(tc.name, func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			f.row(t, "a124-2.9-z2")
			f.led.failEscalate = true
			f.cycles(t, alertAttemptLimit-1)
			a124AtStage(f.d, tc.stage, func() { f.gate.Clear(execgw.ReasonAlertUndelivered) })
			f.cycles(t, 1)
			if f.led.escalationCount() == 0 {
				t.Fatal("no escalation was attempted")
			}
			if !f.latched() {
				t.Fatal("escalation failed and the latch is gone — neither consequence survived")
			}
		})
	}
}

// ㉨/㉧ 오류 판정은 원장의 승인 상태로 버리지 않는다(Y2).
func TestAnErrorJudgementIsNotDroppedByTheLedgersAcknowledgementState(t *testing.T) {
	f := a124Setup(t, &a098RecordingPublisher{})
	id := f.row(t, "a124-2.9-y2")
	other := f.row(t, "a124-2.9-y2-other")
	f.exhaust(t, other, 0)
	f.led.failDeliver = 1
	a124AtStage(f.d, alertStageSettled, func() { _ = f.j.AcknowledgeAlert(context.Background(), id, "operator") })
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — the error hid an acknowledgement but that is no reason to drop it", f.latched(), f.mode(t))
	}
}

// --- 2.10 D8 기록 실패 지속 ------------------------------------------------------------------------

func TestConsecutiveRecordFailuresOnARowLatch(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.10")
	f.led.failRecord = -1
	f.cycles(t, alertAttemptLimit-1)
	if f.latched() {
		t.Fatal("latched before the limit")
	}
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s after %d consecutive record failures", f.latched(), f.mode(t), alertAttemptLimit)
	}
}

func TestClaimFailuresCountAsRecordFailures(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.10-claim")
	f.led.failClaim = -1
	f.cycles(t, alertAttemptLimit)
	if !f.latched() {
		t.Fatal("claim failures on one row never latched")
	}
}

func TestAnotherRowsSuccessDoesNotResetTheCount(t *testing.T) {
	f := a124Setup(t, &a124SelectivePublisher{failTitle: "A-row"})
	ra := f.rowTitled(t, "a124-2.10-a", "A-row")
	f.led.failRecordFor = map[int64]bool{ra: true}
	for i := 0; i < alertAttemptLimit; i++ {
		// 매 사이클 새 B 가 성공해 원장에서 빠진다 — 다른 행의 Applied 가 끼는 사이클.
		f.rowTitled(t, "a124-2.10-b-"+string(rune('a'+i)), "B-row")
		f.cycles(t, 1)
	}
	if !f.latched() {
		t.Fatal("another row's success reset the failing row's record-failure count (R3)")
	}
}

// a124SelectivePublisher 는 제목이 failTitle 인 행만 실패시킨다.
type a124SelectivePublisher struct{ failTitle string }

func (p *a124SelectivePublisher) Publish(_ context.Context, n obs.Notification) error {
	if n.Title == p.failTitle {
		return errors.New("that row cannot be sent")
	}
	return nil
}

func (f *a124Fixture) rowTitled(t *testing.T, key, title string) int64 {
	t.Helper()
	id, err := f.j.EnqueueAlert(context.Background(), journal.Alert{EventKey: key, Type: "x", Severity: "critical", Title: title})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

func TestAClearBetweenRecordFailuresRestartsTheCount(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.10-clear")
	f.led.failRecord = -1
	f.cycles(t, 1)
	f.gate.Clear(execgw.ReasonAlertUndelivered)
	f.cycles(t, alertAttemptLimit-1)
	if f.latched() {
		t.Fatal("a clear request did not break the run of record failures")
	}
	f.cycles(t, 1)
	if !f.latched() {
		t.Fatal("after the clear the run never reached the limit again")
	}
}

func TestAClearAfterTheLimitthRecordFailureStillEscalates(t *testing.T) {
	for _, tc := range []struct {
		name        string
		stage       alertJudgeStage
		cycleBefore bool
	}{
		{"clear after the limit-th error", alertStageSettled, false},
		{"clear between the previous error and the limit-th", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			f.row(t, "a124-2.10-z1")
			f.led.failRecord = -1
			f.cycles(t, alertAttemptLimit-1)
			if tc.cycleBefore {
				f.gate.Clear(execgw.ReasonAlertUndelivered)
			} else {
				a124AtStage(f.d, tc.stage, func() { f.gate.Clear(execgw.ReasonAlertUndelivered) })
			}
			f.cycles(t, 1)
			if f.latched() {
				t.Fatal("latched although a clear followed the last counted error (AA2)")
			}
			if f.mode(t) != journal.ModeEntryBlocked {
				t.Fatalf("mode = %s — at the limit the escalation stands regardless of the clear (Z1)", f.mode(t))
			}
		})
	}
}

func TestLeaseLostAndReleaseDoNotResetTheCount(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.10-lost")
	f.led.failRecord = 1
	f.cycles(t, 1) // 오류 1
	lost := journal.SettleLeaseLost
	f.led.recordOutcome = &lost
	f.cycles(t, 1) // 0 행 기록 + 반납 Applied — 계수 불변
	f.led.recordOutcome = nil
	f.led.failRecord = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit-1)
	if !f.latched() {
		t.Fatal("a lease-lost result or a successful release wiped the record-failure run (R3 · X3)")
	}
}

func TestListingFailuresLatchAfterTheLimit(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.led.failList = -1
	f.cycles(t, alertAttemptLimit-1)
	if f.latched() {
		t.Fatal("latched before the listing limit")
	}
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s after %d listing failures", f.latched(), f.mode(t), alertAttemptLimit)
	}
}

func TestASuccessfulListingBreaksTheListingRun(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.led.failList = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit) // 2 실패 + 1 성공
	f.led.failList = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit-1)
	if f.latched() {
		t.Fatal("a successful listing did not break the run")
	}
}

func TestACancelledEngineIsNotALedgerFault(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.10-cancel")
	f.led.failRecord = -1
	f.led.faultErr = context.Canceled
	listed, err := f.j.PendingAlerts(context.Background(), 0)
	if err != nil || len(listed) != 1 {
		t.Fatalf("arrangement: %v %v", listed, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := 0; i < alertAttemptLimit+1; i++ {
		_ = f.d.cycle(ctx)             // 나열이 취소로 실패 — 세지 않는다
		f.d.deliverOne(ctx, listed[0]) // 임차 · 기록이 취소로 실패 — 세지 않는다
	}
	if f.latched() {
		t.Fatal("a shutdown's cancelled writes were counted as a ledger fault")
	}
}

// --- 2.11 정제 ----------------------------------------------------------------------------------

func TestNewLinesAndTheGateDetailCarryNoSecrets(t *testing.T) {
	const sentinel = "SENTINEL-7731"
	f := a124Setup(t, &a098RecordingPublisher{fail: errors.New("transport " + sentinel)})
	f.d.AccountRef = "acct-" + sentinel
	if _, err := f.j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: "a124-2.11", Type: "x", Severity: "critical",
		Title: "title " + sentinel, Body: "body " + sentinel, Payload: `{"account":"` + sentinel + `"}`,
	}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	f.led.failEscalate = true
	f.led.faultErr = errors.New("ledger " + sentinel)
	f.cycles(t, alertAttemptLimit)
	if !f.latched() {
		t.Fatal("arrangement: not latched")
	}
	if d := f.gate.Blocks()[execgw.ReasonAlertUndelivered]; strings.Contains(d, sentinel) {
		t.Fatalf("gate detail leaks: %q", d)
	}
	for _, line := range strings.Split(f.logs.String(), "\n") {
		if !strings.Contains(line, alertJudgementLogMarker) {
			continue
		}
		if strings.Contains(line, sentinel) {
			t.Fatalf("a new line leaks: %s", line)
		}
	}
	if !strings.Contains(f.logs.String(), alertJudgementLogMarker) {
		t.Fatal("no judgement line was written — the scan measured nothing")
	}
}

// --- 2.12 전달 정산 실패 -------------------------------------------------------------------------

func TestAPublishedButUnrecordedAlertLatchesAtOnceAndKeepsTheLease(t *testing.T) {
	pub := &a098RecordingPublisher{}
	f := a124Setup(t, pub)
	id := f.row(t, "a124-2.12")
	f.led.failDeliver = 1
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s after a publish whose delivery write failed", f.latched(), f.mode(t))
	}
	row, _ := f.j.LookupAlert(context.Background(), id)
	if row.ClaimedBy == "" {
		t.Fatal("the lease was let go — the next cycle would resend an alert that already went out")
	}
	f.cycles(t, 1)
	if n := pub.count(); n != 1 {
		t.Fatalf("published %d times while the lease was live, want 1", n)
	}
	// 만료 뒤의 재발행은 허용된다(R6) — 여기서는 잠그지 않고 기록만.
	f.clk.Advance(journal.DefaultAlertLease + time.Second)
	f.cycles(t, 1)
	if n := pub.count(); n != 2 {
		t.Fatalf("after the lease expired the row was published %d times, want the permitted resend (2)", n)
	}
}

func TestADeliveryRecordOutcomeDecidesTheJudgement(t *testing.T) {
	for _, tc := range []struct {
		out   journal.SettleOutcome
		latch bool
	}{
		{journal.SettleNotFound, true},
		{journal.SettleOutcome(98), true},
		{journal.SettleAlreadySettled, false},
		{journal.SettleLeaseLost, false},
	} {
		t.Run(tc.out.String(), func(t *testing.T) {
			f := a124Setup(t, &a098RecordingPublisher{})
			f.row(t, "a124-2.12-"+tc.out.String())
			o := tc.out
			f.led.deliverOutcome = &o
			f.cycles(t, 1)
			if f.latched() != tc.latch {
				t.Fatalf("%s: latched=%v, want %v", tc.out, f.latched(), tc.latch)
			}
			if tc.latch && f.mode(t) != journal.ModeEntryBlocked {
				t.Fatalf("%s: mode=%s, want escalation", tc.out, f.mode(t))
			}
		})
	}
}

// --- 뮤테이션이 요구한 직접 시험 (tasks 4.1) ------------------------------------------------------

// 판정 입력은 커밋된 값이다(D1, F1): 나열이 0 이라 말한 뒤 다른 발송자가 두 번 실패를 기록하면, 실행자의
// 첫 실패 기록이 커밋한 3 으로 곧장 한도 판정이 선다. 「나열 값 + 1」로 판정하면 1 이라 서지 않는다.
func TestAStaleListingDoesNotDelayTheJudgement(t *testing.T) {
	f := a124Setup(t, a124Failing())
	id := f.row(t, "a124-4.1-stale")
	f.led.beforeClaim = func(int64) { f.exhaust(t, id, alertAttemptLimit-1) }
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — the judgement read the stale listing instead of the committed attempts", f.latched(), f.mode(t))
	}
}

// ㉬ 의 사이클 안 판: 발행 도중의 해제는 근거(실패 기록 커밋) **앞**이다 → 잠금. 세대를 정산 전에 읽으면
// 이 해제가 「뒤」로 보여 판정을 버린다.
type a124ClearingPublisher struct{ gate *execgw.EntryGate }

func (p *a124ClearingPublisher) Publish(context.Context, obs.Notification) error {
	p.gate.Clear(execgw.ReasonAlertUndelivered)
	return errors.New("down")
}

func TestAClearDuringThePublishIsBeforeTheEvidence(t *testing.T) {
	f := a124Setup(t, nil)
	id := f.row(t, "a124-4.1-publish-clear")
	f.exhaust(t, id, alertAttemptLimit-1)
	f.d.Publisher = &a124ClearingPublisher{gate: f.gate}
	f.cycles(t, 1)
	if !f.latched() {
		t.Fatal("a clear that came before the failed attempt was committed dropped the judgement — the epoch was read too early")
	}
}

// Y3: 임차가 「이미 정산됨」을 돌려주면 그 행의 기록 실패 연속은 끝난다. 같은 id 가 새 에피소드로 돌아와도
// 옛 연속을 물려받지 않는다(이 경우는 PENDING 이탈이 관측됐다 — V5 의 해제 없는 재무장과 다르다).
func TestAClaimThatFindsTheRowSettledEndsItsRun(t *testing.T) {
	f := a124Setup(t, a124Failing())
	alert := journal.Alert{EventKey: "a124-4.1-settled", Type: "execgw.order_unresolved", Severity: "critical", Title: "t"}
	id, err := f.j.EnqueueAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	f.led.failRecord = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit-1) // 연속 2
	f.led.beforeClaim = func(int64) {
		if err := f.j.AcknowledgeAlert(context.Background(), id, "operator"); err != nil {
			t.Errorf("AcknowledgeAlert: %v", err)
		}
	}
	f.cycles(t, 1) // 나열에는 있었으나 임차가 「이미 정산됨」
	// 같은 id 를 새 에피소드로 되살린다(재무장) — 다음 나열 전에.
	f.clk.Advance(2 * time.Hour)
	if got := a124State(t, f, id); got != journal.AlertAcknowledged {
		t.Fatalf("arrangement: row is %s, want ACKNOWLEDGED before the re-arm", got)
	}
	again, err := f.j.ClaimAlertForDelivery(context.Background(), alert, time.Hour, "a124-rearm")
	if err != nil || again.ID != id || again.Disposition != journal.ClaimAcquired || again.Stole {
		t.Fatalf("re-arm: %+v %v", again, err)
	}
	if res, err := f.j.ReleaseAlertClaim(context.Background(), id, again.Token); err != nil || res.Outcome != journal.SettleApplied {
		t.Fatalf("release: %+v %v", res, err)
	}
	f.led.failRecord = 1
	f.cycles(t, 1)
	if f.latched() {
		t.Fatal("a row whose claim found it settled kept its old record-failure run into the new episode")
	}
}

// Q5: 잘린 나열에 없다는 것은 PENDING 이탈의 증거가 아니다 — 거기서 계수를 지우면 한도 층으로 밀린 행의
// 연속이 사라진다.
func TestATruncatedListingDoesNotForgetARun(t *testing.T) {
	f := a124Setup(t, &a124SelectivePublisher{failTitle: "A-row"})
	f.d.Batch = 1
	a := f.rowTitled(t, "a124-4.1-trunc-a", "A-row")
	f.exhaust(t, a, alertAttemptLimit) // A 는 한도 층
	f.led.failRecordFor = map[int64]bool{a: true}
	f.cycles(t, alertAttemptLimit-1) // A 만 나열 — 기록 실패 연속 2
	f.rowTitled(t, "a124-4.1-trunc-b", "B-row")
	f.cycles(t, 1) // 나열 = [B] (잘림), B 는 전달됨
	f.cycles(t, 1) // 나열 = [A] — 기록 실패 3 번째
	if !f.latched() {
		t.Fatal("a truncated listing wiped the run of a row it merely did not show")
	}
}

// 세대는 근거(실패 기록 커밋) **뒤**에 읽는다: 기록 직전의 해제는 근거 앞이므로 잠근다. 세대를 기록 전에
// 읽으면 이 해제가 「뒤」로 보여 판정을 버린다(4.1 「세대 읽기를 정산 전으로」).
func TestAClearJustBeforeTheRecordIsBeforeTheEvidence(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-4.1-record-clear")
	f.cycles(t, alertAttemptLimit-1)
	f.led.beforeRecord = func(int64) { f.gate.Clear(execgw.ReasonAlertUndelivered) }
	f.cycles(t, 1)
	if !f.latched() {
		t.Fatal("a clear just before the failed attempt was committed dropped the judgement — the epoch was read before the evidence")
	}
}

// --- codex 구현 1회차 반영 (I1 · I3 · 2.1/2.3/2.8/2.11 보강) ---------------------------------------

// I1: 전달 정산 오류 뒤 취소가 와도 판정은 선다. 취소된 ctx 의 승격 쓰기는 실패하고 → 무조건 차단.
// 승인이 먼저 와서 행이 PENDING 이 아니게 된 판(기동 복원이 덮지 못하는 모양)도 같다.
func TestADeliveryRecordErrorFollowedByCancellationStillLatches(t *testing.T) {
	for _, acknowledged := range []bool{false, true} {
		t.Run(map[bool]string{false: "pending row", true: "row acknowledged first"}[acknowledged], func(t *testing.T) {
			f := a124Setup(t, &a098RecordingPublisher{})
			id := f.row(t, "a124-i1")
			listed, err := f.j.PendingAlerts(context.Background(), 0)
			if err != nil || len(listed) != 1 {
				t.Fatalf("arrangement: %v %v", listed, err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f.led.failDeliver = 1
			f.led.beforeDeliver = func(int64) {
				if acknowledged {
					_ = f.j.AcknowledgeAlert(context.Background(), id, "operator")
				}
				cancel()
			}
			f.d.deliverOne(ctx, listed[0])
			if !f.latched() {
				t.Fatal("a delivery-record error followed by cancellation produced neither latch nor escalation")
			}
		})
	}
}

// ㉫ 판정 대기 중 남이 전달 → 운영자가 빈 backlog 를 승인 · 해제 → 차단 없음, 승격은 남음.
func TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock(t *testing.T) {
	f := a124Setup(t, a124Failing())
	id := f.row(t, "a124-2.9-x2")
	f.cycles(t, alertAttemptLimit-1)
	a124AtStage(f.d, alertStageEpochRead, func() {
		ctx := context.Background()
		claim, err := f.j.ClaimAlertByID(ctx, id, "a124-other-sender")
		if err != nil || claim.Disposition != journal.ClaimAcquired {
			t.Errorf("other sender claim: %+v %v", claim, err)
			return
		}
		if res, err := f.j.MarkAlertDelivered(ctx, id, claim.Token); err != nil || res.Outcome != journal.SettleApplied {
			t.Errorf("other sender delivery: %+v %v", res, err)
		}
		f.acknowledgeAll(t) // 빈 목록 승인 → Clear
	})
	f.cycles(t, 1)
	if got := a124State(t, f, id); got != journal.AlertDelivered {
		t.Fatalf("arrangement: row state %s, want DELIVERED", got)
	}
	if f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — want no block (manual clear after delivery) and the escalation kept", f.latched(), f.mode(t))
	}
}

func a124State(t *testing.T, f *a124Fixture, id int64) string {
	t.Helper()
	row, err := f.j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	return row.State
}

// ㉩ 옛 승인의 늦은 해제(셈~해제 경합의 해제 반쪽)가 근거 뒤에 온다 → 차단 없음 · 승격.
func TestAStaleClearAfterTheEvidenceDropsOnlyTheBlock(t *testing.T) {
	for _, delivered := range []bool{false, true} { // true = ㉪: 그 사이 남이 B 를 전달
		t.Run(map[bool]string{false: "㉩", true: "㉪"}[delivered], func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			id := f.row(t, "a124-2.9-v2")
			f.cycles(t, alertAttemptLimit-1)
			fired := false
			a124AtStage(f.d, alertStageEpochRead, func() {
				fired = true
				if delivered {
					ctx := context.Background()
					claim, err := f.j.ClaimAlertByID(ctx, id, "a124-other-sender")
					if err != nil || claim.Disposition != journal.ClaimAcquired {
						t.Errorf("㉪ arrangement: other sender claim %+v %v", claim, err)
						return
					}
					if res, err := f.j.MarkAlertDelivered(ctx, id, claim.Token); err != nil || res.Outcome != journal.SettleApplied {
						t.Errorf("㉪ arrangement: other sender delivery %+v %v", res, err)
						return
					}
					if got := a124State(t, f, id); got != journal.AlertDelivered {
						t.Errorf("㉪ arrangement: row is %s before the clear, want DELIVERED", got)
					}
				}
				f.gate.Clear(execgw.ReasonAlertUndelivered) // 옛 승인의 늦은 해제
			})
			f.cycles(t, 1)
			if !fired {
				t.Fatal("the stage hook never fired — the schedule was not exercised")
			}
			if f.latched() || f.mode(t) != journal.ModeEntryBlocked {
				t.Fatalf("latched=%v mode=%s", f.latched(), f.mode(t))
			}
		})
	}
}

// ㉢ 도중에 실패한 승인 — 운영자의 승인이 한 행을 승인한 뒤 다음 행에서 원장 오류로 멈춰 해제에 닿지 못한다.
// 해제 요청이 없었으므로 세대는 그대로이고 판정은 선다(차단 + 승격). 실패는 시험 전용 트리거로 **실제로** 일으킨다.
func a124FailAcknowledgementOf(t *testing.T, db *sql.DB, id int64) {
	t.Helper()
	if _, err := db.Exec(fmt.Sprintf(`CREATE TRIGGER a124_fail_ack BEFORE UPDATE OF state ON alert_outbox
		WHEN NEW.id = %d AND NEW.state = 'ACKNOWLEDGED' BEGIN SELECT RAISE(ABORT, 'a124 synthetic acknowledgement failure'); END`, id)); err != nil {
		t.Fatalf("ack-failure trigger: %v", err)
	}
}

func TestAnAcknowledgementThatFailedBeforeItsClearDoesNotDropTheJudgement(t *testing.T) {
	f := a124Setup(t, a124Failing())
	a := f.row(t, "a124-2.9-failed-ack-a")
	f.cycles(t, alertAttemptLimit-1)
	o := f.row(t, "a124-2.9-failed-ack-o") // A 뒤에 생겨 이번 사이클에는 한도 아래 — 자기 판정으로 결과를 가리지 않는다
	side := a124Side(t, f)
	fired := false
	var epochBefore uint64
	a124AtStage(f.d, alertStageEpochRead, func() {
		fired = true
		epochBefore = f.gate.ClearEpoch(execgw.ReasonAlertUndelivered)
		a124FailAcknowledgementOf(t, side, o)
		err := (&obs.Notifier{Journal: f.j, Gate: f.gate}).Acknowledge(context.Background(), "operator")
		if err == nil {
			t.Errorf("the acknowledgement was supposed to fail on row %d", o)
		}
	})
	f.cycles(t, 1)
	if !fired {
		t.Fatal("the stage hook never fired")
	}
	if got := a124State(t, f, a); got != journal.AlertAcknowledged {
		t.Fatalf("arrangement: row A is %s, want ACKNOWLEDGED (the approval got that far)", got)
	}
	if got := f.gate.ClearEpoch(execgw.ReasonAlertUndelivered); got != epochBefore {
		t.Fatalf("the failed acknowledgement moved the clear epoch %d → %d", epochBefore, got)
	}
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — an approval that never reached its clear dropped the judgement", f.latched(), f.mode(t))
	}
}

// 같은 실패한 승인은 기록 실패 연속도 지우지 않는다(Q3).
func TestAFailedAcknowledgementKeepsTheRecordFailureRun(t *testing.T) {
	f := a124Setup(t, a124Failing())
	// O 는 먼저 생겨(승인 순서가 먼저) 다른 발송자가 살아 있는 임차로 쥐고 있다 — 실행자는 건너뛰므로 O 가 자기
	// 판정으로 결과를 가리지 않고, 승인은 임차를 무시하므로 O 는 승인된다.
	o := f.row(t, "a124-2.10-failed-ack-o")
	if claim, err := f.j.ClaimAlertByID(context.Background(), o, "a124-other-sender"); err != nil || claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("arrangement: hold O: %+v %v", claim, err)
	}
	a := f.row(t, "a124-2.10-failed-ack-a")
	f.led.failRecordFor = map[int64]bool{a: true}
	f.cycles(t, alertAttemptLimit-1)
	if run := f.d.recordRuns[a]; run.count != alertAttemptLimit-1 {
		t.Fatalf("arrangement: run = %+v", run)
	}
	a124FailAcknowledgementOf(t, a124Side(t, f), a)
	if err := (&obs.Notifier{Journal: f.j, Gate: f.gate}).Acknowledge(context.Background(), "operator"); err == nil {
		t.Fatal("the acknowledgement was supposed to fail on row A")
	}
	if got := a124State(t, f, o); got != journal.AlertAcknowledged {
		t.Fatalf("arrangement: row O is %s, want ACKNOWLEDGED", got)
	}
	if run := f.d.recordRuns[a]; run.count != alertAttemptLimit-1 {
		t.Fatalf("the failed acknowledgement touched the run: %+v", run)
	}
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — a failed approval wiped the record-failure run", f.latched(), f.mode(t))
	}
}

// AA2 를 나열 계수에도: 한도째 나열 오류 뒤의 해제 · 직전 오류와 한도째 사이의 해제 → 차단 없음 · 승격.
func TestAClearAroundTheLimitthListingFailureStillEscalates(t *testing.T) {
	for _, between := range []bool{false, true} {
		t.Run(map[bool]string{false: "after the limit-th", true: "between"}[between], func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			f.led.failList = -1
			f.cycles(t, alertAttemptLimit-1)
			if between {
				f.gate.Clear(execgw.ReasonAlertUndelivered)
			} else {
				a124AtStage(f.d, alertStageEpochRead, func() { f.gate.Clear(execgw.ReasonAlertUndelivered) })
			}
			f.cycles(t, 1)
			if f.latched() {
				t.Fatal("latched although a clear followed the last counted listing failure")
			}
			if f.mode(t) != journal.ModeEntryBlocked {
				t.Fatalf("mode = %s — the escalation stands at the limit", f.mode(t))
			}
		})
	}
}

// V5: 해제 없이 전달 → 재무장(같은 id)이 실행자의 두 나열 사이에 일어나면 계수는 id 에 남는다 — 새 에피소드가
// 더 일찍 잠길 수 있다(보수 방향, 기대값으로 고정).
func TestARearmWithoutAClearCarriesTheRecordFailureRun(t *testing.T) {
	f := a124Setup(t, a124Failing())
	alert := journal.Alert{EventKey: "a124-2.10-carry", Type: "execgw.order_unresolved", Severity: "critical", Title: "t"}
	id, err := f.j.EnqueueAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	f.led.failRecord = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit-1) // 연속 2
	ctx := context.Background()
	claim, err := f.j.ClaimAlertByID(ctx, id, "a124-other-sender")
	if err != nil || claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("other claim: %+v %v", claim, err)
	}
	if res, err := f.j.MarkAlertDelivered(ctx, id, claim.Token); err != nil || res.Outcome != journal.SettleApplied {
		t.Fatalf("other delivery: %+v %v", res, err)
	}
	// 전달이 **실제로** 있었다 — 없으면 이 행은 PENDING 인 채 임차 만료로 다시 잡혀 재무장 없이도 통과한다(codex 3회차 T1).
	if got := a124State(t, f, id); got != journal.AlertDelivered {
		t.Fatalf("arrangement: row is %s after the other sender's delivery, want DELIVERED", got)
	}
	f.clk.Advance(2 * time.Hour)
	again, err := f.j.ClaimAlertForDelivery(ctx, alert, time.Hour, "a124-rearm")
	if err != nil || again.ID != id || again.Disposition != journal.ClaimAcquired || again.Stole {
		t.Fatalf("re-arm: %+v %v — want a fresh acquisition of a re-armed row, not a steal", again, err)
	}
	if row, err := f.j.LookupAlert(ctx, id); err != nil || row.State != journal.AlertPending || row.Attempts != 0 || row.DeliveredAt != nil {
		t.Fatalf("re-arm did not open a new episode: %+v %v", row, err)
	}
	if res, err := f.j.ReleaseAlertClaim(ctx, id, again.Token); err != nil || res.Outcome != journal.SettleApplied {
		t.Fatalf("release: %+v %v", res, err)
	}
	f.led.failRecord = 1
	f.cycles(t, 1)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s — the re-armed episode did not inherit the run (V5 expectation: latch + escalation on the 3rd error)", f.latched(), f.mode(t))
	}
}

// 2.1 응답 없는 publisher — transport 의 기한 초과도 실패 시도다.
type a124TimeoutPublisher struct{}

func (a124TimeoutPublisher) Publish(context.Context, obs.Notification) error {
	time.Sleep(5 * time.Millisecond)
	return context.DeadlineExceeded
}

func TestAPublisherThatTimesOutCountsAsFailedAttempts(t *testing.T) {
	f := a124Setup(t, a124TimeoutPublisher{})
	f.row(t, "a124-2.1-timeout")
	f.cycles(t, alertAttemptLimit)
	if !f.latched() || f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("latched=%v mode=%s after %d timed-out publishes", f.latched(), f.mode(t), alertAttemptLimit)
	}
}

// 2.3 실제로 원장을 닫고 다시 연 재시작.
func TestARestartFromAReopenedLedgerRelatches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, journal.DBFileName)
	clk := clock.NewFake(time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))
	open := func() *journal.Journal {
		j, err := journal.Open(context.Background(), journal.Options{Path: path, Clock: clk,
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
		if err != nil {
			t.Fatalf("journal.Open: %v", err)
		}
		return j
	}
	j := open()
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	if _, err := j.EnqueueAlert(context.Background(), journal.Alert{EventKey: "a124-2.3-reopen", Type: "x", Severity: "critical"}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	d := &alertDeliverer{Journal: j, Publisher: a124Failing(), Clock: clk, Claimant: "a124", Gate: gate, AccountRef: a124Account}
	for i := 0; i < alertAttemptLimit; i++ {
		_ = d.cycle(context.Background())
	}
	if err := j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	j2 := open()
	t.Cleanup(func() { _ = j2.Close() })
	fresh := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	if err := restoreAlertEntryLatch(context.Background(), j2, fresh); err != nil {
		t.Fatalf("restoreAlertEntryLatch: %v", err)
	}
	if _, ok := fresh.Blocks()[execgw.ReasonAlertUndelivered]; !ok {
		t.Fatal("after a real close/reopen the gate is open with the exhausted row PENDING")
	}
	snap, err := j2.CurrentOperatingMode(context.Background(), a124Account)
	if err != nil || snap.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode row after reopen = %+v %v", snap, err)
	}
}

// 2.8 실제 LeaseLost: 발행 도중 임차가 만료되고 다른 발송자가 가져간다 → 실행자의 기록은 0 행 → 판정 없음.
type a124StealingPublisher struct {
	f  *a124Fixture
	id int64
}

func (p *a124StealingPublisher) Publish(context.Context, obs.Notification) error {
	p.f.clk.Advance(journal.DefaultAlertLease + time.Second)
	claim, err := p.f.j.ClaimAlertByID(context.Background(), p.id, "a124-thief")
	if err != nil || claim.Disposition != journal.ClaimAcquired || !claim.Stole {
		p.f.t.Errorf("steal: %+v %v", claim, err)
	}
	return errors.New("down")
}

func TestARealLeaseLossNeverJudges(t *testing.T) {
	f := a124Setup(t, nil)
	f.t = t
	id := f.row(t, "a124-2.8-steal")
	f.exhaust(t, id, alertAttemptLimit-1)
	f.d.Publisher = &a124StealingPublisher{f: f, id: id}
	f.cycles(t, 1)
	if f.latched() || f.led.escalationCount() != 0 {
		t.Fatalf("a lease lost to another sender still judged: latched=%v escalations=%d", f.latched(), f.led.escalationCount())
	}
}

// 2.11 임차 토큰 sentinel: 발행 뒤 정산 실패로 임차를 쥔 채 둔 행의 실제 토큰이 어느 로그 줄에도 없다.
func TestNoLineCarriesTheLeaseToken(t *testing.T) {
	f := a124Setup(t, &a098RecordingPublisher{})
	id := f.row(t, "a124-2.11-token")
	f.led.failDeliver = 1
	f.cycles(t, 1)
	var token string
	if err := a124Side(t, f).QueryRow(`SELECT claim_token FROM alert_outbox WHERE id = ?`, id).Scan(&token); err != nil || token == "" {
		t.Fatalf("reading the held token: %q %v", token, err)
	}
	if strings.Contains(f.logs.String(), token) {
		t.Fatal("a log line carries the lease token")
	}
	if strings.Contains(f.gate.Blocks()[execgw.ReasonAlertUndelivered], token) {
		t.Fatal("the gate detail carries the lease token")
	}
}

// a124Side 는 같은 원장 파일의 둘째 연결(읽기 전용 확인용).
func a124Side(t *testing.T, f *a124Fixture) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+f.path+"?_pragma=busy_timeout(10000)")
	if err != nil {
		t.Fatalf("side handle: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// R2(codex 2회차): 취소 중 전달 정산 오류의 판정이 조건부 차단을 해제에 뺏겨도, 취소된 승격 쓰기가 실패해
// **무조건 차단**으로 대신한다 — 승격 시도는 있었고 모드 전이는 커밋되지 않았다.
func TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch(t *testing.T) {
	f := a124Setup(t, &a098RecordingPublisher{})
	f.row(t, "a124-r2")
	listed, err := f.j.PendingAlerts(context.Background(), 0)
	if err != nil || len(listed) != 1 {
		t.Fatalf("arrangement: %v %v", listed, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f.led.failDeliver = 1
	f.led.beforeDeliver = func(int64) { cancel() }
	fired := false
	a124AtStage(f.d, alertStageEpochRead, func() {
		fired = true
		f.gate.Clear(execgw.ReasonAlertUndelivered) // 조건부 차단을 뺏는다
	})
	f.d.deliverOne(ctx, listed[0])
	if !fired {
		t.Fatal("the stage hook never fired")
	}
	if f.led.escalationCount() == 0 {
		t.Fatal("no escalation was attempted")
	}
	if f.mode(t) != journal.ModeNormal {
		t.Fatalf("mode = %s — a cancelled escalation write should not have committed", f.mode(t))
	}
	if !f.latched() {
		t.Fatal("the cancelled escalation left no fallback latch")
	}
}

// D9(Eng F4 · codex 2회차): 래치 줄은 이 실행자가 래치를 **새로** 세웠을 때만 — 이미 서 있던 래치(동기 경로 ·
// 기동 복원)의 재확인은 조용하다.
func TestTheLatchLineIsWrittenOnlyWhenThisExecutorLatches(t *testing.T) {
	count := func(f *a124Fixture) int {
		n := 0
		for _, line := range strings.Split(f.logs.String(), "\n") {
			if strings.Contains(line, alertJudgementLogMarker) && strings.Contains(line, string(execgw.ReasonAlertUndelivered)) &&
				strings.Contains(line, string(obs.EventAlertUndelivered)) {
				n++
			}
		}
		return n
	}
	pre := a124Setup(t, a124Failing())
	pre.gate.Block(execgw.ReasonAlertUndelivered, "latched by the synchronous path")
	pre.row(t, "a124-d9-pre")
	pre.cycles(t, alertAttemptLimit+2)
	if n := count(pre); n != 0 {
		t.Fatalf("wrote %d latch lines for a latch this executor did not set", n)
	}
	own := a124Setup(t, a124Failing())
	own.row(t, "a124-d9-own")
	own.cycles(t, alertAttemptLimit+2)
	if n := count(own); n != 1 {
		t.Fatalf("wrote %d latch lines, want exactly 1 for the one latch this executor set", n)
	}
}

// --- gstack /review testing 전문가 반영 -------------------------------------------------------------

// 그 행의 실패 기록이 **적용**되면(attempts 가 오름) 기록 실패 연속은 끝난다 — 연속이 아닌 오류를 이어 세면 D8 의
// 「연속」보다 일찍 잠근다.
func TestAnAppliedRecordOnTheRowEndsItsRun(t *testing.T) {
	f := a124Setup(t, a124Failing())
	id := f.row(t, "a124-review-applied")
	f.led.failRecord = 1
	f.cycles(t, 1) // 기록 오류 1
	if run := f.d.recordRuns[id]; run.count != 1 {
		t.Fatalf("arrangement: run = %+v", run)
	}
	f.cycles(t, 1) // 실제로 적용된 실패 기록(attempts 1) — 연속이 끝나야 한다
	if _, ok := f.d.recordRuns[id]; ok {
		t.Fatalf("an applied record left the run in place: %+v", f.d.recordRuns[id])
	}
	f.led.failRecord = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit-1)
	if f.latched() || f.led.escalationCount() != 0 {
		t.Fatalf("latched=%v escalations=%d — errors separated by an applied record were counted as consecutive", f.latched(), f.led.escalationCount())
	}
}

// Q5 의 양의 절반: 완전한 나열에 없는 행(PENDING 을 떠남)의 연속은 지워진다 — 실행자가 **관측한** 에피소드 경계를
// 넘어 이월되지 않는다(해제 없는 재무장이 두 나열 **사이**에 일어난 V5 와 대비).
func TestACompleteListingForgetsARowThatLeftPending(t *testing.T) {
	f := a124Setup(t, a124Failing())
	alert := journal.Alert{EventKey: "a124-review-prune", Type: "execgw.order_unresolved", Severity: "critical", Title: "t"}
	id, err := f.j.EnqueueAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	f.led.failRecord = alertAttemptLimit - 1
	f.cycles(t, alertAttemptLimit-1) // 연속 2
	ctx := context.Background()
	claim, err := f.j.ClaimAlertByID(ctx, id, "a124-other-sender")
	if err != nil || claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("other claim: %+v %v", claim, err)
	}
	if res, err := f.j.MarkAlertDelivered(ctx, id, claim.Token); err != nil || res.Outcome != journal.SettleApplied {
		t.Fatalf("other delivery: %+v %v", res, err)
	}
	f.cycles(t, 1) // 완전한 나열(행 없음) — 연속이 지워져야 한다
	if _, ok := f.d.recordRuns[id]; ok {
		t.Fatalf("a complete listing without the row kept its run: %+v", f.d.recordRuns[id])
	}
	f.clk.Advance(2 * time.Hour)
	again, err := f.j.ClaimAlertForDelivery(ctx, alert, time.Hour, "a124-rearm")
	if err != nil || again.ID != id || again.Disposition != journal.ClaimAcquired || again.Stole {
		t.Fatalf("re-arm: %+v %v", again, err)
	}
	if res, err := f.j.ReleaseAlertClaim(ctx, id, again.Token); err != nil || res.Outcome != journal.SettleApplied {
		t.Fatalf("release: %+v %v", res, err)
	}
	f.led.failRecord = 1
	f.cycles(t, 1)
	if f.latched() {
		t.Fatal("an episode the executor saw end still carried its record-failure run")
	}
}

// 한도째 판정 뒤 계수는 0 으로 돌아간다(W2) — 다음 오류 하나가 곧장 다시 판정하지 않는다.
func TestTheRunRestartsAfterItsJudgement(t *testing.T) {
	t.Run("record run", func(t *testing.T) {
		f := a124Setup(t, a124Failing())
		id := f.row(t, "a124-review-restart-record")
		f.led.failRecord = -1
		f.cycles(t, alertAttemptLimit)
		if !f.latched() {
			t.Fatal("arrangement: not latched")
		}
		f.gate.Clear(execgw.ReasonAlertUndelivered)
		f.cycles(t, 1)
		if f.latched() {
			t.Fatal("one error after a judgement re-judged at once")
		}
		if run := f.d.recordRuns[id]; run.count != 1 {
			t.Fatalf("run after the judgement = %+v, want a fresh run of 1", run)
		}
	})
	t.Run("listing run", func(t *testing.T) {
		f := a124Setup(t, a124Failing())
		f.led.failList = -1
		f.cycles(t, alertAttemptLimit)
		if !f.latched() {
			t.Fatal("arrangement: not latched")
		}
		f.gate.Clear(execgw.ReasonAlertUndelivered)
		f.cycles(t, 1)
		if f.latched() {
			t.Fatal("one listing error after a judgement re-judged at once")
		}
		if f.d.listRun.count != 1 {
			t.Fatalf("listing run after the judgement = %+v, want a fresh run of 1", f.d.listRun)
		}
	})
}
