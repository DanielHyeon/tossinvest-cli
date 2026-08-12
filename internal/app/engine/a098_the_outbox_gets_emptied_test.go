package engine

// a098 R1 — 기록된 critical 알림은 발송된다.
//
// 오늘 그것을 하는 주체가 없다. Notifier.Flush 는 프로덕션 호출자가 0 이고(a098 §1.4),
// 2판 설계는 그 함수를 부르지도 않는다(design D1.1) — 잠금을 전송 위에서 쥐기 때문이다.
// 그래서 parkAlert 가 쓴 행은 원장에 앉아 있고 아무도 안 집는다.
//
// 관측을 "유한 시간 안에"로 쓰면 안 된다 — 임의로 긴 Sleep 도 유한하다(5라운드 B-P4).
// 가짜 시계에서 **주기 × 2 안에** 로 못 박는다.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a098RecordingPublisher is a live transport that remembers what it was asked to
// send. It is deliberately not a failing one: R1 asks whether a *working* engine
// delivers, and every other R covers the failures.
type a098RecordingPublisher struct {
	mu   sync.Mutex
	sent []obs.Notification
	fail error
}

func (p *a098RecordingPublisher) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail != nil {
		return p.fail
	}
	p.sent = append(p.sent, n)
	return nil
}

func (p *a098RecordingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.sent)
}

// a098ParkedAlert writes the row parkAlert writes.
//
// It goes through EnqueueAlert because that is what parkAlert itself calls
// (execgw/replay.go:551) — the same row, without dragging the replay path in.
func a098ParkedAlert(t *testing.T, j *journal.Journal, key string) int64 {
	t.Helper()
	id, err := j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: key,
		Type:     "execgw.order_unresolved",
		Severity: "critical",
		Title:    "UNRESOLVED_IN_DOUBT: 005930",
		Body:     "an attempt is unresolved",
	})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

// a098AlertState reads one row's state back out of the ledger.
//
// LookupAlert rather than PendingAlerts: the state being asserted is DELIVERED,
// and a listing that only returns PENDING rows cannot tell "delivered" from
// "gone" (outbox.go:518 filters on state).
func a098AlertState(t *testing.T, j *journal.Journal, id int64) string {
	t.Helper()
	row, err := j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert(%d): %v", id, err)
	}
	return row.State
}

// a098RunDeliverer starts the executor and returns a stop function.
func a098RunDeliverer(t *testing.T, d *alertDeliverer) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("the delivery executor did not return after cancellation")
		}
	})
	return cancel, done
}

// TestARecordedCriticalAlertIsDeliveredWithinTwoCycles is R1.
func TestARecordedCriticalAlertIsDeliveredWithinTwoCycles(t *testing.T) {
	j := openTestJournal(t)
	pub := &a098RecordingPublisher{}
	fake := clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))

	id := a098ParkedAlert(t, j, "a098-r1-delivered")

	d := &alertDeliverer{
		Journal:   j,
		Publisher: pub,
		Clock:     fake,
		Interval:  alertDeliveryInterval,
		Batch:     alertDeliveryBatch,
		Claimant:  "a098-test",
	}
	a098RunDeliverer(t, d)

	// 주기 × 2 가 상한이다. 매 전진 전에 sleeper 를 기다리는 것이 "가짜 시계가
	// 실제로 그 주기만큼만 갔다"를 보장한다 — Advance 를 그냥 부르면 루프가 아직
	// 첫 주기를 돌고 있을 수도 있어서 몇 주기가 지났는지 알 수 없다.
	for cycle := 0; cycle < 2; cycle++ {
		if a098AlertState(t, j, id) == journal.AlertDelivered {
			break
		}
		if !fake.WaitForSleepers(1, 5*time.Second) {
			t.Fatalf("cycle %d: the executor never slept, so it is not a periodic loop", cycle)
		}
		fake.Advance(alertDeliveryInterval)
	}

	// 마지막 주기가 끝나기를 기다린다.
	deadline := time.Now().Add(5 * time.Second)
	for a098AlertState(t, j, id) != journal.AlertDelivered && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	if got := a098AlertState(t, j, id); got != journal.AlertDelivered {
		t.Fatalf("alert %d is %s after two cycles, want %s — nobody sent what the outbox kept",
			id, got, journal.AlertDelivered)
	}
	if n := pub.count(); n != 1 {
		t.Errorf("the publisher was asked %d times, want exactly 1", n)
	}
}

// TestTheDelivererHoldsNoLeaseAfterASuccessfulSend pins a099's settlement rule at
// this seam: MarkAlertDelivered releases the lease as part of settling, so the
// executor must not release it again — and must not leave it held either.
func TestTheDelivererHoldsNoLeaseAfterASuccessfulSend(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	pub := &a098RecordingPublisher{}
	fake := clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	id := a098ParkedAlert(t, j, "a098-r1-lease")

	d := &alertDeliverer{
		Journal: j, Publisher: pub, Clock: fake,
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}
	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle: %v", err)
	}

	row, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if row.State != journal.AlertDelivered {
		t.Fatalf("state = %s, want %s", row.State, journal.AlertDelivered)
	}
	if row.ClaimedBy != "" || row.ClaimExpiresAt != nil {
		t.Errorf("the row still carries a lease after settlement: by=%q expires=%v",
			row.ClaimedBy, row.ClaimExpiresAt)
	}
}

// TestTheDelivererNeverPublishesTheSameRowTwiceInACycle is the exclusion R19
// depends on, observed at the cheapest seam: one row, two cycles, one send.
//
// ⛔ 이 테스트가 재는 기전을 정확히 적는다 — 첫 판의 주석은 틀렸다.
// 첫 판은 *"정산을 안 알아채면 다시 보낸다"*고 적었는데, 뮤테이션으로 정산을
// 통째로 건너뛰어도 **이 테스트는 통과했다.** 두 번째 발송을 막은 것은 정산이 아니라
// **살아 있는 임차**다: 주기는 2s 이고 임차는 81s 이므로, 다음 주기의 claim 은
// 같은 claimant 여도 ClaimHeldElsewhere 로 물러난다(a099 의 alertClaimable 은
// 청구자 자신을 특별대우하지 않는다).
//
// 그래서 이 테스트가 지는 것은 **"한 주기 안에서 배제가 선다"** 이고,
// **"정산이 실제로 일어났다"** 는 TestTheDelivererHoldsNoLeaseAfterASuccessfulSend
// 이 진다 — 그쪽이 뮤테이션에 FAIL 했다. 둘을 한 문장으로 적으면 안 된다.
func TestTheDelivererNeverPublishesTheSameRowTwiceInACycle(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	pub := &a098RecordingPublisher{}
	fake := clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	a098ParkedAlert(t, j, "a098-r1-once")

	d := &alertDeliverer{
		Journal: j, Publisher: pub, Clock: fake,
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}
	for i := range 2 {
		if err := d.cycle(ctx); err != nil {
			t.Fatalf("cycle %d: %v", i, err)
		}
	}
	if n := pub.count(); n != 1 {
		t.Errorf("the publisher was asked %d times across two cycles, want 1", n)
	}
}

// TestTheDelivererReturnsPromptlyOnCancellation is R7's cheap half, and it is
// here rather than only in the runtime test because the property belongs to this
// loop: 주기 대기가 취소 가능해야 한다.
//
// The fake clock is never advanced. A loop whose wait ignores ctx would sit in
// Sleep forever and this would time out; a loop using an uncancellable
// time.Sleep would return only after a real 2 seconds.
func TestTheDelivererReturnsPromptlyOnCancellation(t *testing.T) {
	j := openTestJournal(t)
	fake := clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	d := &alertDeliverer{
		Journal: j, Publisher: &a098RecordingPublisher{}, Clock: fake,
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()

	if !fake.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the executor never slept")
	}
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v on cancellation; a clean stop is not a failure", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return within 2s of cancellation — the cycle wait is not cancellable")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Run took %v to return; the wait is finite but not cancellable", elapsed)
	}
}
