package engine_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type blockingRefusalAlerts struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (a *blockingRefusalAlerts) Notify(ctx context.Context, event obs.Event) error {
	if event.Type != obs.EventExitJudgementRefused {
		return nil
	}
	a.once.Do(func() { close(a.entered) })
	select {
	case <-a.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type signallingSubmitter struct {
	delegate *fakeSubmitter
	placed   chan struct{}
	once     sync.Once
}

func (s *signallingSubmitter) Place(ctx context.Context, request execgw.PlaceRequest) (execgw.Outcome, error) {
	outcome, err := s.delegate.Place(ctx, request)
	s.once.Do(func() { close(s.placed) })
	return outcome, err
}

func (s *signallingSubmitter) Cancel(ctx context.Context, request execgw.CancelRequest) (execgw.Outcome, error) {
	return s.delegate.Cancel(ctx, request)
}

func TestCorruptPositionAlertCannotDelayAnotherEmergencyExit(t *testing.T) {
	alerts := &blockingRefusalAlerts{entered: make(chan struct{}), release: make(chan struct{})}
	submitter := &signallingSubmitter{placed: make(chan struct{})}
	h := newExitHarness(t, func(options *engine.ExitObserverOptions) {
		options.Alerts = alerts
		options.Submit = submitter
	})
	submitter.delegate = h.submit

	corruptPosition := h.entry("005930", "10", "10000", "9800", "10000")
	emergencyPosition := h.entry("000660", "10", "10000", "9800", "10000")
	h.prices.last[corruptPosition.Symbol] = 10050
	h.prices.last[emergencyPosition.Symbol] = 10050
	if cycle := h.observer.ObserveOnce(context.Background()); cycle.Err != nil {
		t.Fatalf("priming cycle: %+v", cycle)
	}

	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`UPDATE exit_states SET effective_snapshot_json=effective_snapshot_json || '{}'
		WHERE position_id=?`, corruptPosition.ID); err != nil {
		t.Fatalf("corrupting fixture: %v", err)
	}

	h.prices.last[corruptPosition.Symbol] = 10050
	h.prices.last[emergencyPosition.Symbol] = 9700
	done := make(chan engine.ExitCycle, 1)
	go func() { done <- h.observer.ObserveOnce(context.Background()) }()

	select {
	case <-submitter.placed:
		// The healthy emergency position reached the broker-facing seam first.
	case <-alerts.entered:
		close(alerts.release)
		t.Fatal("corrupt-position alert blocked before the healthy emergency exit was submitted")
	case <-time.After(3 * time.Second):
		close(alerts.release)
		t.Fatal("neither emergency submission nor corruption alert arrived")
	}
	select {
	case <-alerts.entered:
		close(alerts.release)
	case <-time.After(3 * time.Second):
		t.Fatal("corrupt position was not alerted after the emergency submission")
	}
	select {
	case cycle := <-done:
		if cycle.Proposed != 1 {
			t.Fatalf("cycle = %+v, want exactly one healthy emergency proposal", cycle)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cycle did not finish after alert delivery resumed")
	}
}

func TestUnknownLegacyPolicyIdentityIsDurablyGenerationQuarantined(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "10000", "9800", "10000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: "default_v1",
		EntryPrice: "10000", InitialStop: "9800",
	}); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`UPDATE exit_states SET policy_id='UNKNOWN_LEGACY', snapshot_status=NULL,
		policy_version=NULL, policy_digest=NULL, snapshot_id=NULL, decision_id=NULL,
		observation_id=NULL, position_generation=NULL, next_target=NULL, next_protection=NULL,
		last_observation_source=NULL, last_observed_at=NULL, snapshot_action=NULL,
		snapshot_ratio=NULL, projected_quantity=NULL, state_only=NULL, suppressed_reason=NULL,
		effective_snapshot_json=NULL WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 10100)
	cycle := h.observe()
	if cycle.Proposed != 0 || len(h.submit.places) != 0 {
		t.Fatalf("unknown legacy identity reached order path: %+v", cycle)
	}
	q, active, err := h.journal.ActiveExitSnapshotQuarantine(context.Background(), p.ID, p.InstanceSeq)
	if err != nil || !active || q.Reason != "legacy_policy_identity_unknown" || q.Evidence == "" {
		t.Fatalf("generation quarantine = %+v active=%v err=%v", q, active, err)
	}
}
