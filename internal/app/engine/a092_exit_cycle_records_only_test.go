package engine_test

// a092 22.3 C1 (k1 · k2) 행동 시험: NORMAL 계정 + 원격 전송이 멈춘 publisher 에서 exit 관측 사이클이 publish 없이 반환하고,
// 모드 전이 통지와 critical 알림이 outbox 의 PENDING 행으로 남음(발송은 배달 실행자 몫).
//
// 배선(Context.ExitObserver 가 이 인스턴스들을 주입하는지)은 a092_exit_record_only_wiring_test.go 가 재고, 여기서는 그 인스턴스를
// 받은 루프의 행동을 잼. k3(청산 상한 조회의 401)은 같은 Retrier.Query 경로이며 floor 가 exit Retrier 를 쓰는지는 배선 시험이 잼.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// a092StuckPublisher 는 멈춘 원격 전송 — 부르면 release 가 닫힐 때까지 돌아오지 않음(시험 끝에 닫아 누수 없음).
type a092StuckPublisher struct {
	mu      sync.Mutex
	calls   int
	release chan struct{}
}

func (p *a092StuckPublisher) Publish(ctx context.Context, _ obs.Notification) error {
	p.mu.Lock()
	p.calls++
	p.mu.Unlock()
	select {
	case <-p.release:
	case <-ctx.Done():
	}
	return errors.New("stuck transport")
}

func (p *a092StuckPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// a092RecordOnlyHarness 는 exit 하네스에 기록 전용 알림 · 통지자를 꽂음 — 알림기는 멈춘 publisher 와 하네스의 원장 · 게이트를 씀.
func a092RecordOnlyHarness(t *testing.T) (*exitHarness, *a092StuckPublisher) {
	t.Helper()
	pub := &a092StuckPublisher{release: make(chan struct{})}
	t.Cleanup(func() { close(pub.release) })
	var n *obs.Notifier
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		n = &obs.Notifier{
			Publisher: pub, Journal: o.Journal, Gate: o.Retrier.Gate, AccountRef: o.AccountRef,
			Clock: clock.System(), Attempts: 1, RetryDelay: time.Millisecond,
		}
		ro := obs.RecordOnly{N: n}
		o.Alerts = ro
		o.Announcer = ro
		o.Retrier.Escalate = o.Journal
		o.Retrier.AccountRef = o.AccountRef
		o.Retrier.Announcer = ro
	})
	return h, pub
}

// a092ObserveWithin 는 사이클이 d 안에 반환하는지 봄 — 넘으면 루프가 원격 전송을 기다린 것.
func a092ObserveWithin(t *testing.T, h *exitHarness, d time.Duration) engine.ExitCycle {
	t.Helper()
	done := make(chan engine.ExitCycle, 1)
	go func() { done <- h.observer.ObserveOnce(context.Background()) }()
	select {
	case c := <-done:
		return c
	case <-time.After(d):
		t.Fatalf("the exit cycle did not return within %s — it is waiting on the transport", d)
		return engine.ExitCycle{}
	}
}

func a092PendingOfType(t *testing.T, j *journal.Journal, typ string) []journal.Alert {
	t.Helper()
	rows, err := j.PendingAlerts(context.Background(), 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	var out []journal.Alert
	for _, r := range rows {
		if r.Type == typ {
			out = append(out, r)
		}
	}
	return out
}

// k1: 관측 두절 강화 — 사이클이 반환하고, 두절 알림과 모드 통지가 임차 없는 PENDING 행으로 남음.
func TestA092OutageTighteningRecordsWithoutSending(t *testing.T) {
	h, pub := a092RecordOnlyHarness(t)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	a092ObserveWithin(t, h, 10*time.Second)

	h.prices.err = errors.New("down")
	h.clk.Advance(61 * time.Second)
	cycle := a092ObserveWithin(t, h, 10*time.Second)
	if !cycle.Escalated {
		t.Fatal("past the threshold the loop must tighten the operating mode")
	}
	if h.mode() != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s, want ENTRY_BLOCKED", h.mode())
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0 — the exit goroutine must not send", pub.count())
	}
	for _, typ := range []string{string(obs.EventOperatingMode), string(obs.EventExitObservationOutage)} {
		rows := a092PendingOfType(t, h.journal, typ)
		if len(rows) != 1 {
			t.Errorf("%s pending rows = %d, want 1", typ, len(rows))
			continue
		}
		if rows[0].ClaimedBy != "" {
			t.Errorf("%s row carries a lease by %q", typ, rows[0].ClaimedBy)
		}
	}
}

// k2: 가격 조회 401 — 게이트 래치 · 모드 강화 · 통지 행, publish 없음.
func TestA092CredentialTighteningRecordsWithoutSending(t *testing.T) {
	h, pub := a092RecordOnlyHarness(t)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.prices.err = fmt.Errorf("price read: %w", official.ErrAuth)

	a092ObserveWithin(t, h, 10*time.Second)
	if h.mode() != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s, want ENTRY_BLOCKED after a rejected credential", h.mode())
	}
	if rej := h.gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonBrokerAuthRejected {
		t.Errorf("entry check = %v, want %s", rej, execgw.ReasonBrokerAuthRejected)
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0", pub.count())
	}
	if rows := a092PendingOfType(t, h.journal, string(obs.EventOperatingMode)); len(rows) != 1 {
		t.Errorf("mode announcement rows = %d, want 1", len(rows))
	}
}
