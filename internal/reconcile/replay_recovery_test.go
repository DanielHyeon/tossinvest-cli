package reconcile_test

// Recovery call-graph tests (extend-execution-contract task 2.3).
//
// The requirement is that *this* package owns the order — "재시작 복구가 호출
// 순서를 소유한다" — and that the replay door stays on the Gateway while the
// resolver keeps its no-writer invariant. Both are properties of the wiring, so
// both are tested against the wiring rather than asserted in a comment.

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// fakeReplayer stands in for the gateway's replay entry point. It records what
// it was asked to replay, which is how the tests below prove the ordering.
type fakeReplayer struct {
	j *journal.Journal
	// settleAs, when terminal, is what the replay does to the attempt before
	// reporting. An empty value means "the replay established nothing".
	settleAs journal.AttemptState
	orderID  string
	err      error
	calls    []string
}

func (f *fakeReplayer) ReplayInDoubt(ctx context.Context, attemptID string) (execgw.ReplayOutcome, error) {
	f.calls = append(f.calls, attemptID)
	if f.err != nil {
		return execgw.ReplayOutcome{AttemptID: attemptID}, f.err
	}
	out := execgw.ReplayOutcome{AttemptID: attemptID, State: journal.StateInDoubt, QueryFallback: true}
	if !f.settleAs.IsTerminal() {
		out.Reason = execgw.ReasonReplayExhausted
		return out, nil
	}

	attempt, err := f.j.Resume(ctx, attemptID)
	if err != nil {
		return out, err
	}
	switch f.settleAs {
	case journal.StateConfirmed:
		if err := attempt.ResolveConfirmed(ctx, f.orderID,
			journal.ReasonReplayRecovered, "recovered by replay"); err != nil {
			return out, err
		}
		out.State = journal.StateConfirmed
		out.BrokerOrderID = f.orderID
		out.Reason = execgw.ReasonReplayRecovered
	case journal.StateUnresolvedInDoubt:
		if err := attempt.ResolveUnresolved(ctx, journal.ReasonReplayKeyConflict,
			"the key names a different order"); err != nil {
			return out, err
		}
		out.State = journal.StateUnresolvedInDoubt
		out.Reason = execgw.ReasonReplayKeyConflict
	}
	out.QueryFallback = false
	return out, nil
}

// TestRecoveryReplaysBeforeObserving: the replay recovers the identity, and the
// observation procedure is never asked. Without the replay this attempt would
// have been declared *absent* — the fixture broker shows nothing — so the two
// paths disagree, which is what makes the ordering observable.
func TestRecoveryReplaysBeforeObserving(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	crashMidDispatch(t, path)

	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	opts := recoveryOptions(j, gate, recoveryCollector(nil, nil))
	replayer := &fakeReplayer{j: j, settleAs: journal.StateConfirmed, orderID: "O-recovered"}
	opts.Replayer = replayer

	r, err := reconcile.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(replayer.calls) != 1 || replayer.calls[0] != "attempt-crash" {
		t.Fatalf("replayed attempts: got %v, want [attempt-crash]", replayer.calls)
	}
	if len(report.Resolutions) != 0 {
		t.Fatalf("the observation procedure ran for an attempt the replay settled: %+v", report.Resolutions)
	}
	if len(report.Replays) != 1 || report.Replays[0].State != journal.StateConfirmed {
		t.Fatalf("replays = %+v, want one CONFIRMED", report.Replays)
	}
	stored, err := j.LookupAttempt(context.Background(), "attempt-crash")
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != journal.StateConfirmed || stored.BrokerOrderID != "O-recovered" {
		t.Fatalf("journal: state=%s brokerOrderID=%q, want the replayed identity",
			stored.State, stored.BrokerOrderID)
	}
}

// TestRecoveryFallsBackToObservationWhenReplayDeclines is the other half of the
// call graph. A declined replay is not an error and does not stop the sequence:
// the query fallback is what it falls back *to*.
func TestRecoveryFallsBackToObservationWhenReplayDeclines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	crashMidDispatch(t, path)

	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	opts := recoveryOptions(j, gate, recoveryCollector(nil, nil))
	replayer := &fakeReplayer{j: j} // establishes nothing
	opts.Replayer = replayer

	r, err := reconcile.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(replayer.calls) != 1 {
		t.Fatalf("replay calls: got %d, want 1", len(replayer.calls))
	}
	if len(report.Replays) != 1 || !report.Replays[0].QueryFallback {
		t.Fatalf("replays = %+v, want one asking for the fallback", report.Replays)
	}
	if len(report.Resolutions) != 1 {
		t.Fatalf("resolutions = %+v, want the observation procedure to have taken over", report.Resolutions)
	}
	if report.Resolutions[0].State != journal.StateFailedConfirmed {
		t.Fatalf("resolution state = %s, want FAILED_CONFIRMED (proven absent by observation)",
			report.Resolutions[0].State)
	}
}

// TestRecoveryRecordsAReplayThatParked: a parked attempt is settled, so the
// resolver is not asked — but it still has to appear in the report's Unresolved
// list, because that is what an operator reads.
func TestRecoveryRecordsAReplayThatParked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	crashMidDispatch(t, path)

	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	opts := recoveryOptions(j, gate, recoveryCollector(nil, nil))
	opts.Replayer = &fakeReplayer{j: j, settleAs: journal.StateUnresolvedInDoubt}

	r, err := reconcile.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Unresolved) != 1 || report.Unresolved[0] != "attempt-crash" {
		t.Fatalf("unresolved = %v, want the parked attempt", report.Unresolved)
	}
	if len(report.Resolutions) != 0 {
		t.Fatalf("a parked attempt must not be re-run through observation: %+v", report.Resolutions)
	}
}

// TestRecoveryWithoutAReplayerObservesOnly pins the default this build ships:
// no replayer wired, so the sequence is exactly what it was before replay
// existed.
func TestRecoveryWithoutAReplayerObservesOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	crashMidDispatch(t, path)

	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	r, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil)))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Replays) != 0 {
		t.Fatalf("replays = %+v, want none", report.Replays)
	}
	if len(report.Resolutions) != 1 {
		t.Fatalf("resolutions = %+v, want the observation procedure alone", report.Resolutions)
	}
}

// TestReplayErrorStopsTheSequence: recovery completing over an attempt whose
// replay blew up would release the entry latch without having established
// anything about a possibly-live order.
func TestReplayErrorStopsTheSequence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	crashMidDispatch(t, path)

	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	opts := recoveryOptions(j, gate, recoveryCollector(nil, nil))
	opts.Replayer = &fakeReplayer{j: j, err: errors.New("the journal is unreadable")}

	r, err := reconcile.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Run(context.Background()); !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("Run: got %v, want ErrRecoveryIncomplete", err)
	}
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Fatal("an incomplete recovery must leave the entry latch shut")
	}
}

// TestTheReplayDoorIsOnTheGatewayNotTheResolver is the P1 invariant this change
// had to preserve: the resolution procedure has no writer of any kind, so the
// replay entry point could not live on it. The check is structural because the
// invariant is — a resolver that could replay would be a resolver that could put
// a request on the wire, and no amount of care at the call sites would take that
// back.
func TestTheReplayDoorIsOnTheGatewayNotTheResolver(t *testing.T) {
	replayer := reflect.TypeOf((*reconcile.Replayer)(nil)).Elem()
	if !reflect.TypeOf((*execgw.Gateway)(nil)).Implements(replayer) {
		t.Error("the gateway must be the replay entry point")
	}
	if reflect.TypeOf((*execgw.Resolver)(nil)).Implements(replayer) {
		t.Error("the resolver implements the replay entry point; the no-writer invariant is gone")
	}

	resolver := reflect.TypeOf(execgw.Resolver{})
	forbidden := map[reflect.Type]string{
		reflect.TypeOf((*trading.Service)(nil)):               "a trading service",
		reflect.TypeOf((*execgw.Gateway)(nil)):                "a gateway",
		reflect.TypeOf((*execgw.ReplayTransport)(nil)).Elem(): "a replay transport",
	}
	for i := 0; i < resolver.NumField(); i++ {
		field := resolver.Field(i)
		if what, bad := forbidden[field.Type]; bad {
			t.Errorf("Resolver.%s is %s; the resolution procedure must hold no writer", field.Name, what)
		}
	}
}
