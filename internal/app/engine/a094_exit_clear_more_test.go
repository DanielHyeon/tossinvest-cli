package engine_test

// a094 §3 나머지 — 3.R4(막힌 전송자 · 실제 알림기) · 3.R9 재시작(같은 행) · 4.N4f②③ · 3.9 · 3.D1 · 3.7 · 3.5.

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a094BlockedPublisher 는 원격 전송이 영영 돌아오지 않는 전송자임.
type a094BlockedPublisher struct{ calls int }

func (p *a094BlockedPublisher) Publish(ctx context.Context, _ obs.Notification) error {
	p.calls++
	<-ctx.Done()
	return ctx.Err()
}

func (h *exitHarness) a094Rows(fragment string) []journal.Alert {
	h.t.Helper()
	rows, err := h.journal.PendingAlerts(context.Background(), 1000)
	if err != nil {
		h.t.Fatalf("PendingAlerts: %v", err)
	}
	var out []journal.Alert
	for _, r := range rows {
		if strings.Contains(r.EventKey, fragment) {
			out = append(out, r)
		}
	}
	return out
}

// 3.R4 — 새 critical 은 실제 알림기의 기록 입구로 창 0 적재만 함: 전송자가 막혀도 관측 사이클은 기다리지 않고 다른 포지션의
// 손절이 같은 사이클에 나감. 전송자는 한 번도 불리지 않음.
func TestA094ABlockedSenderDoesNotHoldAnotherStop(t *testing.T) {
	pub := &a094BlockedPublisher{}
	var notifier *obs.Notifier
	h, sub, _ := a094Harness(t, func(o *engine.ExitObserverOptions) {
		notifier = &obs.Notifier{Journal: o.Journal, Publisher: pub}
		o.Critical = notifier
	})
	parked := a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
	other := h.entry("000660", "10", "70000", "68000", "70000")
	h.quote("000660", 70100)
	h.observe()
	places := len(h.submit.places)

	h.quote("005930", 68500) // 첫 포지션: park 원인 critical
	h.quote("000660", 67000) // 둘째 포지션: 손절
	done := make(chan struct{})
	go func() { h.observe(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the observation cycle waited on the blocked sender")
	}
	if len(h.a094Rows("|"+parked.ID+"|attempt:")) != 1 {
		t.Fatal("control: the park-cause row was not recorded")
	}
	if pub.calls != 0 {
		t.Errorf("publisher calls = %d, want 0 — the record entry never sends", pub.calls)
	}
	if len(h.submit.places) != places+1 || h.submit.places[len(h.submit.places)-1].Intent.Symbol != "000660" {
		t.Errorf("the other position's stop was not placed in the same cycle (places %d→%d)", places, len(h.submit.places))
	}
	_ = other
}

// 3.R9 · 4.N4f②③ — 재시작 뒤 같은 에피소드(park attempt · 취소 attempt)는 같은 원장 행, 다른 attempt 는 새 행.
func TestA094AnEpisodeIsOneRowAcrossRestarts(t *testing.T) {
	h, sub, _ := a094Harness(t, func(o *engine.ExitObserverOptions) {
		o.Critical = &obs.Notifier{Journal: o.Journal}
	})
	p := a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
	h.quote("005930", 68500)
	h.observe()
	if n := len(h.a094Rows("|" + p.ID + "|attempt:")); n != 1 {
		t.Fatalf("park rows = %d, want 1", n)
	}
	// 재시작: 같은 원장 위의 새 관측자(관측자 래치 없음).
	restarted, err := engine.NewExitObserver(h.a094Options(t))
	if err != nil {
		t.Fatal(err)
	}
	restarted.ObserveOnce(context.Background())
	if n := len(h.a094Rows("|" + p.ID + "|attempt:")); n != 1 {
		t.Fatalf("park rows after the restart = %d, want the same single row", n)
	}
}

// 3.9 — 엔진에 귀속되지 않는 주문은 취소 대상이 아님(원장에 없으면 목록에 없음) — 오늘 동작 고정.
func TestA094AnOrderTheEngineDidNotPlaceIsNotCancelled(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 67900)
	h.observe()
	if len(h.submit.cancels) != 0 {
		t.Fatalf("cancels = %d with no engine-placed working order, want 0", len(h.submit.cancels))
	}
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d, want the stop", len(h.submit.places))
	}
}

// 3.D1 — 발의가 없는 주기(withPending=false)에 자기 방향 매도가 있어도 보호 청산은 제출된다(오늘 동작).
func TestA094AnOwnSideSellDoesNotWithholdAStopWithoutAProposal(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.a094RecordSell("other-sell", 3, 71000, journal.StateConfirmed)
	h.quote("005930", 67900)
	h.observe()
	if len(h.submit.cancels) != 0 || len(h.submit.places) != 1 {
		t.Fatalf("cancels %d places %d, want the stop submitted and the other sell left alone", len(h.submit.cancels), len(h.submit.places))
	}
}

// 3.7 — 원장 목록 읽기 실패는 종전대로 오류(사이클 오류, 제출 0).
func TestA094AListReadFailureIsAnErrorAsBefore(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`ALTER TABLE fill_snapshots RENAME TO fill_snapshots_gone`); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	h.quote("005930", 67900)
	cycle := h.observe()
	if cycle.Err == nil || len(h.submit.places) != 0 {
		t.Fatalf("cycle err %v places %d, want the read failure surfaced and nothing submitted", cycle.Err, len(h.submit.places))
	}
}

// a094Options 는 하네스와 같은 배선의 옵션을 다시 만듦(재시작 모사).
func (h *exitHarness) a094Options(t *testing.T) engine.ExitObserverOptions {
	t.Helper()
	opts := h.observer.OptionsForTest()
	return opts
}
