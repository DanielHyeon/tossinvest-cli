package reconcile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

type releaseControllingStore struct {
	reconcile.ReconcileStore
	releaseErr error
	released   bool
	requests   []journal.ReleaseReconcileRequest
}

type conflictingEnterStore struct {
	reconcile.ReconcileStore
	existing journal.ReconcileState
}

type retryingEnterStore struct {
	reconcile.ReconcileStore
	remainingFailures int
	calls             int
}

type gateAssertingEnterStore struct {
	reconcile.ReconcileStore
	gate *execgw.EntryGate
}

type sequencedReleaseStore struct {
	reconcile.ReconcileStore
	calls  int
	failAt int
}

type blockingAuthorityReadStore struct {
	reconcile.ReconcileStore
	readStarted chan struct{}
	allowReturn chan struct{}
	enterCalled chan struct{}
}

func (s *blockingAuthorityReadStore) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	states, err := s.ReconcileStore.ActiveReconcileStates(ctx)
	close(s.readStarted)
	<-s.allowReturn
	return states, err
}

func (s *blockingAuthorityReadStore) EnterReconcile(ctx context.Context,
	req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	close(s.enterCalled)
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *sequencedReleaseStore) ReleaseReconcile(ctx context.Context,
	req journal.ReleaseReconcileRequest) (journal.ReconcileState, bool, error) {
	s.calls++
	if s.failAt > 0 && s.calls == s.failAt {
		return journal.ReconcileState{}, false, errors.New("journal temporarily unavailable")
	}
	return s.ReconcileStore.ReleaseReconcile(ctx, req)
}

func (s *gateAssertingEnterStore) EnterReconcile(ctx context.Context,
	req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if s.gate.CheckEntryFor("us", req.Symbol) == nil {
		return journal.ReconcileState{}, false, errors.New("entry gate was open during durable enter")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *retryingEnterStore) EnterReconcile(ctx context.Context,
	req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	s.calls++
	if s.remainingFailures > 0 {
		s.remainingFailures--
		return journal.ReconcileState{}, false, errors.New("journal temporarily unavailable")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *conflictingEnterStore) EnterReconcile(context.Context,
	journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	return s.existing, false, nil
}

func (s *releaseControllingStore) ReleaseReconcile(_ context.Context,
	req journal.ReleaseReconcileRequest) (journal.ReconcileState, bool, error) {
	s.requests = append(s.requests, req)
	return journal.ReconcileState{}, s.released, s.releaseErr
}

func (s *releaseControllingStore) ReleaseReconciles(_ context.Context,
	reqs []journal.ReleaseReconcileRequest) ([]journal.ReconcileState, error) {
	s.requests = append(s.requests, reqs...)
	if s.releaseErr != nil {
		return nil, s.releaseErr
	}
	if !s.released {
		return nil, errors.New("atomic exact-cause release refused")
	}
	return make([]journal.ReconcileState, len(reqs)), nil
}

// restore_test.go is the restart test task 4.1 asks for: 재시작 차단 유지.
//
// A tracker and a gate are built, a disagreement is observed, and then both are
// thrown away — the way a process restart throws away everything in memory. New
// ones are built against the *same journal file*, restored, and asked the same
// question. If the answer changes, the engine would open a position into an
// account it still disagrees with.

func trackerOn(clk clock.Clock, gate *execgw.EntryGate, j *journal.Journal) *reconcile.Tracker {
	return &reconcile.Tracker{
		Clock:       clk,
		Gate:        gate,
		Journal:     j,
		MinInterval: 30 * time.Second,
		MaxFailures: 3,
		AccountRef:  "acct-7",
	}
}

// noStaleGate blocks on latches only; a staleness block would answer every
// question before the reconcile state was consulted.
func noStaleGate(clk clock.Clock) *execgw.EntryGate {
	return execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
}

func TestARestartKeepsTheReconcileBlock(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"

	first := openJournalAt(t, path)
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, first)

	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if tracker.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("precondition: the disagreement must block AAPL")
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("precondition: the gate must carry the block")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// --- restart: everything in memory is gone -------------------------------

	second := openJournalAt(t, path)
	restartedGate := noStaleGate(clk)
	restarted := trackerOn(clk, restartedGate, second)

	if restartedGate.CheckEntryFor("us", "AAPL") != nil {
		t.Fatal("precondition: a fresh gate starts empty, which is what Restore has to fix")
	}
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	rejected := restartedGate.CheckEntryFor("us", "AAPL")
	if rejected == nil {
		t.Fatal("the block was lost across the restart; the journal still carries the disagreement")
	}
	if rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("reason = %s, want %s", rejected.Reason, execgw.ReasonReconcileMismatch)
	}
	if restarted.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("the tracker's own view must be restored too")
	}
	// Unrelated symbols are still tradable: the restore honours the scope.
	if restartedGate.CheckEntryFor("us", "MSFT") != nil {
		t.Fatal("a symbol-scoped block must not survive as an account-wide one")
	}
}

func TestARestartKeepsAPermanentMismatchPermanent(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"

	first := openJournalAt(t, path)
	tracker := trackerOn(clk, noStaleGate(clk), first)
	for i := 0; i < 3; i++ {
		if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
			t.Fatalf("Observe %d: %v", i, err)
		}
		clk.Advance(31 * time.Second)
	}
	if !tracker.Permanent() {
		t.Fatal("precondition: three consecutive failures make the mismatch permanent")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second := openJournalAt(t, path)
	restartedGate := noStaleGate(clk)
	restarted := trackerOn(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if !restarted.Permanent() {
		t.Fatal("a permanent mismatch must survive a restart as permanent")
	}
	if restartedGate.CheckEntry() == nil {
		t.Fatal("a permanent mismatch is account-wide; the account must be blocked")
	}

	// The dangerous case: one clean pass after the restart must not release a
	// block only an operator can clear.
	if _, err := restarted.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched: 1}); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if !restarted.Permanent() {
		t.Fatal("a clean reconciliation must not clear a permanent mismatch")
	}
	if restartedGate.CheckEntry() == nil {
		t.Fatal("the account block must survive a clean pass after a restart")
	}
}

func TestAnOperatorResolutionSurvivesARestart(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"

	first := openJournalAt(t, path)
	tracker := trackerOn(clk, noStaleGate(clk), first)
	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if err := tracker.Resolve(ctx, "daniel", "compared the app to the journal by hand"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second := openJournalAt(t, path)
	restartedGate := noStaleGate(clk)
	restarted := trackerOn(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if rejected := restartedGate.CheckEntryFor("us", "AAPL"); rejected != nil {
		t.Fatalf("a resolved block must not come back on restart: %v", rejected)
	}
	if len(restarted.Blocks()) != 0 {
		t.Fatalf("the tracker restored blocks an operator had cleared: %+v", restarted.Blocks())
	}
}

func TestRefreshCannotOverwriteABlockPersistedByObserve(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	store := &blockingAuthorityReadStore{
		ReconcileStore: j,
		readStarted:    make(chan struct{}),
		allowReturn:    make(chan struct{}),
		enterCalled:    make(chan struct{}),
	}
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	tracker.Journal = store

	refreshDone := make(chan error, 1)
	go func() { refreshDone <- tracker.Refresh(ctx) }()
	<-store.readStarted // the refresh captured an empty durable snapshot

	observeDone := make(chan error, 1)
	go func() {
		_, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4"))
		observeDone <- err
	}()
	select {
	case <-store.enterCalled:
		t.Fatal("Observe wrote through while Refresh still owned an older authority read")
	case <-time.After(50 * time.Millisecond):
		// Expected: Refresh owns the tracker lock until its authority snapshot is
		// installed, so Observe cannot persist yet.
	}
	close(store.allowReturn)
	if err := <-refreshDone; err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if err := <-observeDone; err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Symbol != "AAPL" {
		t.Fatalf("stale refresh overwrote the newly persisted block: %+v", blocks)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("stale refresh reopened the entry gate")
	}
}

// TestRestoreProjectsStatesThisTrackerDidNotEnter: the gate is the projection of
// the *whole* table, not only of this tracker's rows. An identifier conflict
// recorded by the resolution path still has to block after a restart.
func TestRestoreProjectsStatesThisTrackerDidNotEnter(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)

	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-7", Symbol: "AAPL",
		Cause:    journal.ReconcileCauseIdentifierConflict,
		Evidence: "order 42 turned up under two symbols",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}

	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	rejected := gate.CheckEntryFor("us", "AAPL")
	if rejected == nil {
		t.Fatal("a state another producer recorded must still block after a restart")
	}
	if rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("an identifier conflict is operator-only; reason = %s", rejected.Reason)
	}
	// Adoption and the runtime read the tracker projection, so the exact durable
	// cause must be present there too. It remains owned by its original producer:
	// a quantity comparison cannot release it.
	blocks := tracker.Blocks()
	if len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseIdentifierConflict {
		t.Fatalf("tracker projection = %+v, want the exact identifier-conflict state", blocks)
	}
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL")
	clk.Advance(31 * time.Second)
	if _, err := tracker.Observe(ctx, cleanDiffAt(clk)); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if blocks = tracker.Blocks(); len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseIdentifierConflict {
		t.Fatalf("a quantity comparison changed another producer's block: %+v", blocks)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("a clean quantity comparison released an identifier conflict it says nothing about")
	}
}

func TestRestoreProjectsEveryActiveCauseForTheConfiguredAccount(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	states := []journal.EnterReconcileRequest{
		{AccountRef: "acct-7", Symbol: "GOOG", Cause: journal.ReconcileCauseSnapshotUnavailable, Evidence: "holdings unreadable"},
		{AccountRef: "acct-7", Symbol: "AAPL", Cause: journal.ReconcileCauseSnapshotStale, Evidence: "holdings stale"},
		{AccountRef: "acct-7", Symbol: "MSFT", Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "quantity differs"},
		{AccountRef: "acct-7", Symbol: "TSLA", Cause: journal.ReconcileCauseIdentifierConflict, Evidence: "identifier conflict"},
		{AccountRef: "acct-7", Symbol: "NVDA", Cause: journal.ReconcileCauseAttributionFailed, Evidence: "unattributed record"},
		{AccountRef: "other-account", Symbol: "META", Cause: journal.ReconcileCauseIdentifierConflict, Evidence: "foreign account"},
	}
	for _, req := range states {
		if _, _, err := j.EnterReconcile(ctx, req); err != nil {
			t.Fatalf("EnterReconcile(%s): %v", req.Cause, err)
		}
	}

	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	blocks := tracker.Blocks()
	if len(blocks) != 5 {
		t.Fatalf("blocks = %+v, want all five configured-account causes", blocks)
	}
	got := map[string]reconcile.Block{}
	for _, block := range blocks {
		got[block.Cause] = block
	}
	for _, cause := range []string{
		journal.ReconcileCauseSnapshotUnavailable,
		journal.ReconcileCauseSnapshotStale,
		journal.ReconcileCauseQuantityMismatch,
		journal.ReconcileCauseIdentifierConflict,
		journal.ReconcileCauseAttributionFailed,
	} {
		if got[cause].Cause != cause {
			t.Errorf("cause %s missing from tracker projection: %+v", cause, blocks)
		}
	}
	if got[journal.ReconcileCauseQuantityMismatch].Release != reconcile.ReleaseOnAdjustedReconcile {
		t.Errorf("quantity mismatch release = %q, want adjusted reconcile",
			got[journal.ReconcileCauseQuantityMismatch].Release)
	}
	for _, cause := range []string{
		journal.ReconcileCauseSnapshotUnavailable,
		journal.ReconcileCauseSnapshotStale,
		journal.ReconcileCauseIdentifierConflict,
		journal.ReconcileCauseAttributionFailed,
	} {
		if got[cause].Release != reconcile.ReleaseOperatorOnly {
			t.Errorf("other-producer cause %s release = %q, want operator-only in this tracker",
				cause, got[cause].Release)
		}
	}
	if gate.CheckEntryFor("us", "META") != nil {
		t.Fatal("Restore projected another account's state into this account gate")
	}
}

func TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails(t *testing.T) {
	for _, tc := range []struct {
		name       string
		released   bool
		releaseErr error
	}{
		{name: "store error", releaseErr: errors.New("disk unavailable")},
		{name: "store declined exact cause", released: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clk := clock.NewFake(asOf)
			j := openJournal(t)
			if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
				AccountRef: "acct-7", Symbol: "AAPL",
				Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "quantity differs",
			}); err != nil {
				t.Fatalf("EnterReconcile: %v", err)
			}
			gate := noStaleGate(clk)
			tracker := trackerOn(clk, gate, j)
			if err := tracker.Restore(ctx); err != nil {
				t.Fatalf("Restore: %v", err)
			}
			store := &releaseControllingStore{
				ReconcileStore: j, released: tc.released, releaseErr: tc.releaseErr,
			}
			tracker.Journal = store
			tracker.AdjustmentApplied(asOfAt(clk), "AAPL")
			clk.Advance(31 * time.Second)

			out, err := tracker.Observe(ctx, cleanDiffAt(clk))
			if err == nil {
				t.Fatal("Observe must report an uncommitted durable release")
			}
			if len(out.Cleared) != 0 || !out.Blocked {
				t.Fatalf("outcome = %+v, want no visible release and an active block", out)
			}
			if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseQuantityMismatch {
				t.Fatalf("memory block disappeared after failed release: %+v", blocks)
			}
			if gate.CheckEntryFor("us", "AAPL") == nil {
				t.Fatal("gate reopened before the durable release committed")
			}
			if len(store.requests) != 1 || store.requests[0].ExpectCause != journal.ReconcileCauseQuantityMismatch {
				t.Fatalf("release requests = %+v, want exact quantity cause", store.requests)
			}
		})
	}
}

func TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	for _, symbol := range []string{"AAPL", "MSFT"} {
		if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
			AccountRef: "acct-7", Symbol: symbol,
			Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "quantity differs",
		}); err != nil {
			t.Fatalf("EnterReconcile(%s): %v", symbol, err)
		}
	}
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	store := &sequencedReleaseStore{ReconcileStore: j, failAt: 2}
	tracker.Journal = store
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL", "MSFT")
	clk.Advance(31 * time.Second)

	out, err := tracker.Observe(ctx, reconcile.Diff{AsOf: asOfAt(clk), AccountRef: "acct-7", Matched: 2})
	if err == nil {
		t.Fatal("Observe must report the second durable release failure")
	}
	if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "AAPL" {
		t.Fatalf("cleared = %+v, want only the first durable release", out.Cleared)
	}
	if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Symbol != "MSFT" {
		t.Fatalf("memory blocks after partial persistence = %+v, want MSFT only", blocks)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].Symbol != "MSFT" {
		t.Fatalf("journal after partial persistence = %+v, want MSFT only", active)
	}
	if gate.CheckEntryFor("us", "AAPL") != nil || gate.CheckEntryFor("us", "MSFT") == nil {
		t.Fatal("gate did not mirror the exact durable partial-release set")
	}

	store.failAt = 0
	clk.Advance(31 * time.Second)
	out, err = tracker.Observe(ctx, reconcile.Diff{AsOf: asOfAt(clk), AccountRef: "acct-7", Matched: 2})
	if err != nil {
		t.Fatalf("retry Observe: %v", err)
	}
	if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "MSFT" || out.Blocked {
		t.Fatalf("retry outcome = %+v, want the retained adjustment to release MSFT", out)
	}
	if gate.CheckEntryFor("us", "MSFT") != nil {
		t.Fatal("successful retry left the remaining gate closed")
	}
}

func TestResolveKeepsTheRefusedCauseBlocked(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-7", Symbol: "AAPL",
		Cause: journal.ReconcileCauseIdentifierConflict, Evidence: "identifier conflict",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	store := &releaseControllingStore{ReconcileStore: j}
	tracker.Journal = store

	if err := tracker.Resolve(ctx, "daniel", "fresh comparison was clean"); err == nil {
		t.Fatal("Resolve must fail when the store declines the exact-cause release")
	}
	if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseIdentifierConflict {
		t.Fatalf("refused operator release changed memory: %+v", blocks)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("refused operator release reopened the gate")
	}
	if len(store.requests) != 1 || store.requests[0].ExpectCause != journal.ReconcileCauseIdentifierConflict {
		t.Fatalf("release requests = %+v, want exact identifier-conflict cause", store.requests)
	}
}

func TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-7", Cause: journal.ReconcileCauseSnapshotUnavailable,
		Evidence: "account holdings unavailable",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	tracker := trackerOn(clk, noStaleGate(clk), j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	for i := 0; i < 3; i++ {
		if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
			t.Fatalf("Observe %d: %v", i, err)
		}
		clk.Advance(31 * time.Second)
	}
	var accountBlock *reconcile.Block
	for _, block := range tracker.Blocks() {
		if block.Scope == reconcile.ScopeAccount {
			copy := block
			accountBlock = &copy
		}
	}
	if accountBlock == nil || accountBlock.Cause != journal.ReconcileCauseSnapshotUnavailable {
		t.Fatalf("account block = %+v, want the original snapshot-unavailable authority", accountBlock)
	}
	if tracker.Permanent() {
		t.Fatal("a blocked promotion must not claim that a quantity permanent row was persisted")
	}
}

func TestEnterConflictReplacesTheProposalWithTheDurableCause(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, nil)
	tracker.Journal = &conflictingEnterStore{
		existing: journal.ReconcileState{
			AccountRef: "acct-7", Symbol: "AAPL",
			Cause: journal.ReconcileCauseIdentifierConflict, Evidence: "durable conflict",
			EnteredAt: asOf,
		},
	}

	out, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4"))
	if err == nil {
		t.Fatal("a different durable cause at the proposed scope must fail the cycle")
	}
	if !out.Blocked || len(out.Cleared) != 0 {
		t.Fatalf("outcome = %+v, want fail-closed", out)
	}
	blocks := tracker.Blocks()
	if len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseIdentifierConflict {
		t.Fatalf("blocks = %+v, want the returned durable cause, not the quantity proposal", blocks)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("gate = %v, want the returned operator-held cause", rejected)
	}
}

func TestFailedEnterIsRetriedUntilTheBlockIsDurable(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	store := &retryingEnterStore{ReconcileStore: j, remainingFailures: 1}
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, nil)
	tracker.Journal = store
	diff := mismatchDiff("AAPL", "10", "4")

	if _, err := tracker.Observe(ctx, diff); err == nil {
		t.Fatal("first enter must expose the durable write failure")
	}
	if gate.CheckEntryFor("us", "AAPL") == nil || len(tracker.Blocks()) != 1 {
		t.Fatal("failed persistence must keep the in-memory block and gate closed")
	}
	if active, err := j.ActiveReconcileStates(ctx); err != nil || len(active) != 0 {
		t.Fatalf("active after failed enter = %+v, err=%v; want no false durable row", active, err)
	}

	clk.Advance(31 * time.Second)
	if _, err := tracker.Observe(ctx, diff); err != nil {
		t.Fatalf("second observation did not retry the pending enter: %v", err)
	}
	if store.calls != 2 {
		t.Fatalf("EnterReconcile calls = %d, want one initial attempt and one retry", store.calls)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 1 || active[0].Cause != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("active after retry = %+v, want durable quantity block", active)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil || len(tracker.Blocks()) != 1 {
		t.Fatal("durable retry must remain visibly blocked")
	}
}

func TestNewMismatchLatchesTheGateBeforeJournalIO(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, nil)
	tracker.Journal = &gateAssertingEnterStore{ReconcileStore: j, gate: gate}

	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe reached journal I/O before latching the gate: %v", err)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("new mismatch did not remain latched after durable enter")
	}
}

func TestSuccessfulForeignCauseResolutionClearsItsSymbolGate(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-7", Symbol: "AAPL",
		Cause: journal.ReconcileCauseIdentifierConflict, Evidence: "identifier conflict",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("precondition: identifier conflict must block the symbol")
	}

	if err := tracker.Resolve(ctx, "daniel", "fresh stable comparison was clean"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected != nil {
		t.Fatalf("successful exact-cause resolution left a stale symbol gate: %v", rejected)
	}
}

// TestWriteThroughRecordsTheStateWithItsEvidence pins that the journal — not
// the block map — is where the disagreement lives.
func TestWriteThroughRecordsTheStateWithItsEvidence(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	tracker := trackerOn(clk, noStaleGate(clk), j)

	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("want one persisted state, got %+v", active)
	}
	state := active[0]
	if state.Symbol != "AAPL" || state.Cause != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("persisted state = %+v", state)
	}
	if state.Evidence == "" {
		t.Fatal("the persisted state must carry the evidence an operator acts on")
	}

	// A clean pass on its own leaves it standing. Task 6.3: 조정 없이 우연히
	// 일치한 단발 관측(SHALL NOT) — the release needs something to have converged
	// the projection, not just a reading that happened to agree.
	clk.Advance(31 * time.Second)
	if _, err := tracker.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched: 1}); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if active, err = j.ActiveReconcileStates(ctx); err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("a coincidental match must not close the state, got %+v", active)
	}

	// An adjustment and the re-read after it do close it, and the journal says
	// which of the two was the cause. This assertion used to be on the clean pass
	// above, with cause RECHECK_MATCHED.
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL")
	clk.Advance(31 * time.Second)
	if _, err := tracker.Observe(ctx, cleanDiffAt(clk)); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if active, err = j.ActiveReconcileStates(ctx); err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("an adjustment and a matching re-read must close the state, got %+v", active)
	}
	history, err := j.ReconcileStateHistory(ctx, "acct-7")
	if err != nil {
		t.Fatalf("ReconcileStateHistory: %v", err)
	}
	if len(history) != 1 || history[0].ReleaseCause != journal.ReconcileReleaseAdjustmentApplied {
		t.Fatalf("want one state released by ADJUSTMENT_APPLIED, got %+v", history)
	}
}
