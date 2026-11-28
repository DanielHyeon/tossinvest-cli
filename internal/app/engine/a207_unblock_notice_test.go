package engine_test

// a207 — 막힌 청산 재시도의 즉시성 회귀(T1) · 해소 알림(T6) · 해제 셈(M12).

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func a207Refused() *execgw.Outcome {
	return &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected, Detail: "no"}
}

func a207Arms(t *testing.T, h *exitHarness, positionID string) []journal.ExitEvent {
	t.Helper()
	events, err := h.journal.ExitEvents(context.Background(), positionID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	var arms []journal.ExitEvent
	for _, e := range events {
		if e.ProposedIntentID != "" && exitpolicy.Action(e.Action).Orderable() {
			arms = append(arms, e)
		}
	}
	return arms
}

// T1 — 즉시성 회귀: 다섯 주기 거절이 이어져도 주기마다 제출은 정확히 한 번, 막힘이 풀린 첫 주기에 손절이 나가고,
// 증거는 구간 첫 행만 전체 · 나머지는 압축, 해소 알림은 한 번.
func TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.submit.placeOutcome = a207Refused()
	const n = 5
	for i := 0; i < n; i++ {
		before := len(h.submit.places)
		h.clk.Advance(engine.DefaultExitObservationInterval)
		h.quote("005930", 67900-float64(10*i))
		h.observe()
		if got := len(h.submit.places) - before; got != 1 {
			t.Fatalf("cycle %d placed %d orders, want exactly one — a refused stop is proposed again on the next observation", i+1, got)
		}
		if h.state(p.ID).Pending() {
			t.Fatalf("cycle %d: the refused proposal is still armed", i+1)
		}
	}
	if got := h.alerts.count(obs.EventExitProposalUnblocked); got != 0 {
		t.Fatalf("unblocked notices = %d while still refused, want 0", got)
	}

	h.submit.placeOutcome = nil // 막힘이 풀림
	before := len(h.submit.places)
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67850)
	h.observe()
	if got := len(h.submit.places) - before; got != 1 {
		t.Fatalf("the first clear cycle placed %d orders, want the stop submitted on that very cycle", got)
	}

	arms := a207Arms(t, h, p.ID)
	if len(arms) != n+1 {
		t.Fatalf("arm rows = %d, want one per cycle (%d) — compaction must not drop rows", len(arms), n+1)
	}
	if arms[0].Evaluation.EffectiveSource != journal.EffectiveSourceRecomputed || arms[0].Evaluation.Effective.Snapshot == nil {
		t.Fatalf("the first arm = %+v, want full evidence", arms[0].Evaluation)
	}
	for i, a := range arms[1:] {
		if a.Evaluation.EffectiveSource != journal.EffectiveSourceRecomputedRepeat ||
			a.Evaluation.Effective.UnknownReason != journal.UnknownReasonRepeatOfRefusedProposal {
			t.Fatalf("arm %d = %+v, want a compact repeat row", i+2, a.Evaluation)
		}
	}

	if got := h.alerts.count(obs.EventExitProposalUnblocked); got != 1 {
		t.Fatalf("unblocked notices = %d, want one", got)
	}
	notice, _ := h.alerts.first(obs.EventExitProposalUnblocked)
	if obs.SeverityOf(notice.Type) == obs.SeverityCritical {
		t.Fatal("the unblock notice must be normal grade")
	}
	if notice.Fields["refusals"] != n || notice.Fields["position_id"] != p.ID {
		t.Fatalf("notice fields = %+v, want %d refusals for the position", notice.Fields, n)
	}

	// 다음 주기에 다시 나가지 않음.
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67840)
	h.observe()
	if got := h.alerts.count(obs.EventExitProposalUnblocked); got != 1 {
		t.Fatalf("unblocked notices = %d after a further cycle, want still one", got)
	}
}

// T6 — 거절 없는 수용은 해소 알림이 없음.
func TestA207AnAcceptedStopWithoutRefusalsIsNotAnUnblock(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67900)
	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d, want the stop submitted", len(h.submit.places))
	}
	if got := h.alerts.count(obs.EventExitProposalUnblocked); got != 0 {
		t.Fatalf("unblocked notices = %d without a refusal, want 0", got)
	}
}

// a207Seq 는 Place 와 알림의 순서를 한 줄로 적음.
type a207Seq struct{ log []string }

type a207SeqAlerts struct {
	seq    *a207Seq
	events []obs.Event
}

func (a *a207SeqAlerts) Notify(_ context.Context, e obs.Event) error {
	a.seq.log = append(a.seq.log, "alert:"+string(e.Type))
	a.events = append(a.events, e)
	return nil
}

type a207SeqSubmit struct {
	*fakeSubmitter
	seq *a207Seq
}

func (s *a207SeqSubmit) Place(ctx context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	s.seq.log = append(s.seq.log, "place")
	return s.fakeSubmitter.Place(ctx, req)
}

// T6 · M9 — 해소 알림은 주기 끝에 남: 같은 주기의 모든 Place 뒤. 두 포지션이 함께 풀리면 판정 순서와 무관하게
// 둘째 포지션의 Place 가 첫째 알림보다 앞서야 함.
func TestA207UnblockNoticesFollowEveryPlaceOfTheCycle(t *testing.T) {
	seq := &a207Seq{}
	alerts := &a207SeqAlerts{seq: seq}
	var submit *a207SeqSubmit
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		submit = &a207SeqSubmit{fakeSubmitter: o.Submit.(*fakeSubmitter), seq: seq}
		o.Submit = submit
		o.Alerts = alerts
	})
	h.entry("005930", "10", "70000", "68000", "70000")
	h.entry("000660", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.quote("000660", 70100)
	h.observe()
	h.submit.placeOutcome = a207Refused()
	for i := 0; i < 2; i++ {
		h.clk.Advance(engine.DefaultExitObservationInterval)
		h.quote("005930", 67900-float64(10*i))
		h.quote("000660", 67900-float64(10*i))
		h.observe()
	}
	h.submit.placeOutcome = nil
	seq.log = nil
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67800)
	h.quote("000660", 67800)
	h.observe()

	lastPlace, firstNotice, notices := -1, -1, 0
	for i, entry := range seq.log {
		switch entry {
		case "place":
			lastPlace = i
		case "alert:" + string(obs.EventExitProposalUnblocked):
			notices++
			if firstNotice < 0 {
				firstNotice = i
			}
		}
	}
	if notices != 2 {
		t.Fatalf("unblock notices = %d in %v, want one per position", notices, seq.log)
	}
	if lastPlace < 0 || firstNotice < lastPlace {
		t.Fatalf("sequence %v: an unblock notice came before a Place of the same cycle", seq.log)
	}
}

// M12 — 원장이 해제를 거절하면(released=false) 셈을 더하지 않음.
type a207ConfirmedButRefused struct {
	*fakeSubmitter
	record func(execgw.PlaceRequest) string
}

func (s *a207ConfirmedButRefused) Place(_ context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	s.places = append(s.places, req)
	s.record(req) // 원장에는 접수 확정 attempt — 해제 판정이 거절함
	return *a207Refused(), nil
}

func TestA207AReleaseTheLedgerRefusesIsNotCounted(t *testing.T) {
	var submit *a207ConfirmedButRefused
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		submit = &a207ConfirmedButRefused{fakeSubmitter: o.Submit.(*fakeSubmitter)}
		o.Submit = submit
	})
	submit.record = h.recordSubmittedSell
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67900)
	h.observe()
	if len(submit.places) != 1 {
		t.Fatalf("places = %d, want one", len(submit.places))
	}
	if !h.state(p.ID).Pending() {
		t.Fatal("fixture: the ledger was expected to refuse the release and keep the proposal armed")
	}
	if got := h.observer.RefusalStreakForTest(p.ID); got != 0 {
		t.Fatalf("refusal streak = %d after a release the ledger refused, want 0", got)
	}
}

// D7 — 알림기가 없는 조립도 거절 → 수용을 그대로 돎(해소 알림 없음, 패닉 없음).
func TestA207NoAlertsWiredStillSubmitsAfterRefusals(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Alerts = nil })
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.submit.placeOutcome = a207Refused()
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67900)
	h.observe()
	h.submit.placeOutcome = nil
	h.clk.Advance(engine.DefaultExitObservationInterval)
	h.quote("005930", 67890)
	h.observe()
	if len(h.submit.places) != 2 {
		t.Fatalf("places = %d, want the refused stop and the accepted retry", len(h.submit.places))
	}
}
