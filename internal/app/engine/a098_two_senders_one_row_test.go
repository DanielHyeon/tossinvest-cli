package engine

// a098 — 같은 행에 두 주체가 붙었을 때, 그리고 붙은 채로 멈췄을 때.
//
//	R19  배달 실행자와 동기 발송 경로가 같은 행을 **동시에 안 보낸다**
//	R11  정산이 실패해도 **임차를 안 푼다** — 안 그러면 다음 주기가 다시 보낸다
//	R21  매 주기 held 로 남는 행은 **한 번은 보이고, 주기마다 쏟아지지는 않는다**
//
// 셋 다 a099의 임차가 있어야 성립한다. R19는 배제 그 자체이고, R11은 배제가
// **끝나면 안 되는 시점**이며, R21은 배제가 **오래 끌 때** 무엇이 보이는가다.

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a098BlockingPublisher holds whoever is publishing until the test lets go.
//
// 이것이 「동시에」를 결정적으로 만드는 조각이다. 두 goroutine 을 그냥 띄우면
// 겹치는지 아닌지가 운에 달리고, 안 겹친 판은 **아무것도 안 재고 통과한다.**
type a098BlockingPublisher struct {
	entered chan struct{}
	release chan struct{}
	fail    error

	mu   sync.Mutex
	n    int
	once sync.Once
}

func (p *a098BlockingPublisher) Publish(context.Context, obs.Notification) error {
	p.mu.Lock()
	p.n++
	p.mu.Unlock()
	p.once.Do(func() { close(p.entered) })
	<-p.release
	return p.fail
}

func (p *a098BlockingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.n
}

// a098CountingPublisher records how many sends were attempted and can fail them.
type a098CountingPublisher struct {
	mu   sync.Mutex
	n    int
	fail error
	// hook runs inside Publish, before it returns. R11 uses it to make the world
	// change underneath a send that already succeeded.
	hook func()
}

func (p *a098CountingPublisher) Publish(context.Context, obs.Notification) error {
	p.mu.Lock()
	p.n++
	p.mu.Unlock()
	if p.hook != nil {
		p.hook()
	}
	return p.fail
}

func (p *a098CountingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.n
}

// a098CriticalEvent is the condition both senders are looking at.
//
// `Key` 를 명시한다 — 두 경로가 **같은 행**을 보는 것이 이 테스트의 전제이고,
// 파생 키에 맡기면 그 전제가 구현 세부에 달린다.
func a098CriticalEvent(key string) obs.Event {
	return obs.Event{
		Type:  obs.EventOrderInDoubt,
		Key:   key,
		Title: "UNRESOLVED_IN_DOUBT: 005930",
		Body:  "an attempt is unresolved",
	}
}

func a098SyncNotifier(j *journal.Journal, pub obs.Publisher, logs *bytes.Buffer) *obs.Notifier {
	return &obs.Notifier{
		Log:        obs.NewLogger(obs.LogOptions{Writer: logs, JSON: true, Clock: clock.System()}),
		Publisher:  pub,
		Journal:    j,
		AccountRef: "a098-r19",
		Clock:      clock.System(),
		Attempts:   1,
		RetryDelay: time.Millisecond,
	}
}

// TestTheSynchronousSenderStandsDownWhileTheExecutorHoldsTheRow is R19, one way.
//
// 동기 경로가 행을 쥐고 **발송 중일 때** 배달 실행자가 무엇을 하는가. 답은
// 「아무것도」여야 하고, 그 근거는 임차다 — 실행자가 그 순간에도 보내면 운영자는
// 같은 알림을 두 번 받는다. 2026-08-08 이 정확히 그 형태였다.
func TestTheSynchronousSenderStandsDownWhileTheExecutorHoldsTheRow(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	var logs bytes.Buffer

	syncPub := &a098BlockingPublisher{entered: make(chan struct{}), release: make(chan struct{})}
	n := a098SyncNotifier(j, syncPub, &logs)

	notified := make(chan error, 1)
	go func() { notified <- n.Notify(ctx, a098CriticalEvent("a098-r19-sync-holds")) }()

	select {
	case <-syncPub.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the synchronous sender never reached its publisher")
	}

	// 동기 경로가 임차를 쥔 채 전송 중이다. 이 순간 실행자를 돌린다.
	execPub := &a098CountingPublisher{}
	d := &alertDeliverer{
		Journal: j, Publisher: execPub,
		Log:      obs.NewLogger(obs.LogOptions{Writer: &logs, JSON: true, Clock: clock.System()}),
		Clock:    clock.NewFake(a098BatchNow),
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}
	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle while the synchronous sender holds the row: %v", err)
	}
	if got := execPub.count(); got != 0 {
		t.Fatalf("the executor published %d times while the row was being sent; "+
			"운영자가 같은 알림을 두 번 받는다", got)
	}
	if !strings.Contains(logs.String(), "held by another sender") {
		t.Errorf("the executor stood down silently: %s", logs.String())
	}

	close(syncPub.release)
	select {
	case err := <-notified:
		if err != nil {
			t.Fatalf("Notify: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Notify never returned")
	}

	// 정확히 한 번. 둘을 합쳐서 센다 — 한쪽만 보면 다른 쪽의 중복을 못 본다.
	if total := syncPub.count() + execPub.count(); total != 1 {
		t.Errorf("the two senders published %d times between them, want exactly 1", total)
	}
}

// TestTheExecutorHoldsTheRowAndTheSynchronousSenderStandsDown is R19, the other
// way — and it is a separate test because the two can fail independently.
//
// 배제는 **대칭이어야** 한다. 한 방향만 재면 「실행자는 물러나지만 동기 경로는
// 안 물러나는」 구현이 통과하고, 그 구현에서 중복은 그대로 일어난다.
func TestTheExecutorHoldsTheRowAndTheSynchronousSenderStandsDown(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	var logs bytes.Buffer

	// 1. 동기 경로가 행을 **만든다** — 전송은 실패시켜서 PENDING 으로 남긴다.
	//    (실패한 critical 전송은 임차를 돌려준다: notifier.go 의 release 경로)
	syncPub := &a098CountingPublisher{fail: errors.New("the transport is down")}
	n := a098SyncNotifier(j, syncPub, &logs)
	event := a098CriticalEvent("a098-r19-exec-holds")
	_ = n.Notify(ctx, event)
	if syncPub.count() != 1 {
		t.Fatalf("the first notify attempted %d sends, want 1", syncPub.count())
	}

	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil || len(pending) != 1 {
		t.Fatalf("PendingAlerts = %d rows, err=%v — 전제가 안 섰다", len(pending), err)
	}

	// 2. 실행자가 그 행을 쥐고 전송 중에 멈춘다.
	execPub := &a098BlockingPublisher{entered: make(chan struct{}), release: make(chan struct{})}
	d := &alertDeliverer{
		Journal: j, Publisher: execPub,
		Log:      obs.NewLogger(obs.LogOptions{Writer: &logs, JSON: true, Clock: clock.System()}),
		Clock:    clock.NewFake(a098BatchNow),
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}
	cycled := make(chan error, 1)
	go func() { cycled <- d.cycle(ctx) }()
	select {
	case <-execPub.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the executor never reached its publisher")
	}

	// 3. 같은 조건이 다시 관측된다. 동기 경로는 물러나야 한다.
	if err := n.Notify(ctx, event); err != nil {
		t.Fatalf("the second Notify: %v", err)
	}
	if got := syncPub.count(); got != 1 {
		t.Fatalf("the synchronous sender published again (%d attempts) while the executor "+
			"held the row", got)
	}

	close(execPub.release)
	select {
	case err := <-cycled:
		if err != nil {
			t.Fatalf("cycle: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the cycle never finished")
	}
	if got := execPub.count(); got != 1 {
		t.Errorf("the executor published %d times, want 1", got)
	}
	if got := a098AlertState(t, j, pending[0].ID); got != journal.AlertDelivered {
		t.Errorf("state = %s, want %s — 배제만 하고 아무도 안 보냈다", got, journal.AlertDelivered)
	}
}

// TestAFailedSettlementKeepsTheLeaseSoTheNextCycleDoesNotResend is R11.
//
// **2판의 「어느 쪽이든 끝나면 해제」가 깨는 것이 이것이다.** 발송은 성공했는데
// 기록이 실패한 상태에서 임차를 돌려주면, 다음 주기가 그 행을 집어 **이미 나간 것을
// 다시 보낸다** — 성공 경로로 도착하는 2026-08-08 이다.
//
// ⛔ 대조군이 없으면 **죽은 루프도 통과한다.** 두 번째 주기가 실제로 돌았다는 것을
// **다른 행이 나가는 것**으로 증명한다.
func TestAFailedSettlementKeepsTheLeaseSoTheNextCycleDoesNotResend(t *testing.T) {
	j := openTestJournal(t)
	id := a098ParkedAlert(t, j, "a098-r11-unsettled")

	// 발송은 성공하고, 그 직후 정산이 실패하게 만든다: 정산은 주기의 ctx 를 쓰므로
	// Publish 안에서 그 ctx 를 취소하면 **publish 성공 + 정산 실패**가 결정적으로 선다.
	failing, cancelDuringPublish := context.WithCancel(context.Background())
	pub := &a098CountingPublisher{hook: cancelDuringPublish}
	d := &alertDeliverer{
		Journal: j, Publisher: pub, Clock: clock.NewFake(a098BatchNow),
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}
	if err := d.cycle(failing); err != nil {
		t.Fatalf("cycle 1: %v", err)
	}
	if pub.count() != 1 {
		t.Fatalf("cycle 1 published %d times, want 1 — 전제가 안 섰다", pub.count())
	}

	row, err := j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if row.State != journal.AlertPending {
		t.Fatalf("state = %s, want %s — 정산이 실패하지 않았다면 R11 은 공허하다",
			row.State, journal.AlertPending)
	}
	if row.ClaimedBy != "a098-test" || row.ClaimExpiresAt == nil {
		t.Fatalf("the lease was handed back after a failed settlement: by=%q expires=%v — "+
			"다음 주기가 이미 나간 것을 다시 보낸다", row.ClaimedBy, row.ClaimExpiresAt)
	}

	// --- 대조군: 다음 주기가 실제로 돈다 -------------------------------------
	ctx := context.Background()
	control := a098ParkedAlert(t, j, "a098-r11-control")
	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle 2: %v", err)
	}
	if got := a098AlertState(t, j, control); got != journal.AlertDelivered {
		t.Fatalf("the control row is %s after cycle 2; 루프가 안 돌았으므로 "+
			"「다시 안 보낸다」는 아무것도 증명하지 않는다", got)
	}
	if got := pub.count(); got != 2 {
		t.Errorf("the publisher was asked %d times in total, want 2 (the control only) — "+
			"임차를 안 풀었는데도 다시 보냈다", got)
	}
	if got := a098AlertState(t, j, id); got != journal.AlertPending {
		t.Errorf("the unsettled row is %s; 정산은 여전히 안 됐어야 한다", got)
	}
}

// TestAHeldRowIsReportedOncePerLeaseAndNotOncePerCycle is R21.
//
// 두 방향을 다 본다:
//
//	① 같은 임차에서 K 주기를 돌아도 **한 줄**  — 로그 폭풍은 관측이 아니다
//	② 임차가 바뀌면 **다시 한 줄**            — 「한 번 찍고 영원히 침묵」은 은폐다
//
// ①만 보면 아무것도 안 찍는 구현이 통과하고, ②만 보면 매 주기 쏟는 구현이 통과한다.
func TestAHeldRowIsReportedOncePerLeaseAndNotOncePerCycle(t *testing.T) {
	ctx := context.Background()
	const lease = 10 * time.Second
	j, ledgerClock := a098LeasedJournal(t, lease)
	var logs bytes.Buffer

	id := a098ParkedAlert(t, j, "a098-r21-stuck")
	if _, err := j.ClaimAlertByID(ctx, id, "engine.alert_delivery.stalled"); err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}

	d := &alertDeliverer{
		Journal: j, Publisher: &a098CountingPublisher{},
		Log:      obs.NewLogger(obs.LogOptions{Writer: &logs, JSON: true, Clock: ledgerClock}),
		Clock:    clock.NewFake(a098BatchNow),
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}

	const cycles = 6
	for i := range cycles {
		if err := d.cycle(ctx); err != nil {
			t.Fatalf("cycle %d: %v", i, err)
		}
	}
	if got := strings.Count(logs.String(), "held by another sender"); got != 1 {
		t.Fatalf("%d cycles produced %d 'held' lines, want exactly 1 — "+
			"주기마다 한 줄이면 그 다음에 오는 줄을 아무도 안 읽는다", cycles, got)
	}

	// --- ② 임차가 바뀌면 다시 말한다 ------------------------------------------
	//
	// 두 번째 발송자가 만료된 임차를 가져간다. 그것은 **다른 발송자에 대한 새로운
	// 사실**이고, 첫 번째와 뭉뚱그리면 인수 자체가 안 보인다.
	ledgerClock.Advance(lease + time.Second)
	if _, err := j.ClaimAlertByID(ctx, id, "engine.alert_delivery.second"); err != nil {
		t.Fatalf("the second sender could not take the expired lease: %v", err)
	}
	for i := range cycles {
		if err := d.cycle(ctx); err != nil {
			t.Fatalf("cycle %d after the handover: %v", i, err)
		}
	}
	if got := strings.Count(logs.String(), "held by another sender"); got != 2 {
		t.Fatalf("after the handover there are %d 'held' lines in total, want 2 — "+
			"임차가 바뀐 것이 안 보이면 발송자가 바뀐 것도 안 보인다", got)
	}
	if !strings.Contains(logs.String(), "engine.alert_delivery.second") {
		t.Errorf("the new holder is not named: %s", logs.String())
	}
}
