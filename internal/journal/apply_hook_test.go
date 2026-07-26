package journal

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// apply_hook_test.go pins the atomic apply point (task 0.3).
//
// Three things are asserted, and none of the three is enough alone:
//
//  1. behaviour — the injected functions run inside the fill transaction, and a
//     failure in one of them takes the fill with it;
//  2. reachability — the guarded exit columns cannot be written from outside a
//     hook, because the handle that writes them cannot be obtained from outside
//     a hook;
//  3. layout — no other production file in this package writes those columns at
//     all, which is the part a behavioural test can never show (it can only say
//     that today's path does not).

// applyHookFixture opens a journal with a position and an exit state to apply
// against.
func applyHookFixture(t *testing.T) *Journal {
	t.Helper()
	j := openTestJournal(t)
	insertPosition(t, j, "p-1", nil)
	insertExitState(t, j, "p-1")
	return j
}

func takenRatio(t *testing.T, j *Journal, positionID string) string {
	t.Helper()
	var total string
	if err := j.db.QueryRowContext(context.Background(),
		"SELECT taken_ratio_total FROM exit_states WHERE position_id = ?", positionID).
		Scan(&total); err != nil {
		t.Fatalf("reading taken_ratio_total: %v", err)
	}
	return total
}

func armPendingForTest(t *testing.T, j *Journal, positionID string) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(),
		`UPDATE exit_states
		    SET pending_action = 'PARTIAL_TAKE', pending_level = ?, pending_intent_id = 'intent-1'
		  WHERE position_id = ?`, RatchetPartialLock, positionID); err != nil {
		t.Fatalf("arming a pending proposal: %v", err)
	}
}

// TestApplyHooksRunInsideTheFillTransaction is the requirement itself: the
// snapshot, the projection and the exit state commit together. The hooks write
// through the handle, and everything they wrote is visible after RecordFill
// returns — from one commit, not two.
func TestApplyHooksRunInsideTheFillTransaction(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()
	armPendingForTest(t, j, "p-1")

	var order, projectedDelta string
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
			order += "project,"
			projectedDelta = fill.Delta
			_, err := tx.Exec(ctx,
				`UPDATE positions SET quantity = ?, state = ? WHERE id = 'p-1'`,
				fill.CumulativeQuantity, PositionOpen)
			return err
		},
		Exit: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
			order += "exit"
			// The projection's write is visible: same transaction.
			rows, err := tx.Query(ctx, "SELECT quantity FROM positions WHERE id = 'p-1'")
			if err != nil {
				return err
			}
			defer rows.Close()
			if !rows.Next() {
				return errors.New("the exit hook cannot see the projection's write")
			}
			var quantity string
			if err := rows.Scan(&quantity); err != nil {
				return err
			}
			if quantity != fill.CumulativeQuantity {
				return errors.New("the exit hook saw a stale projection")
			}
			rows.Close()

			if err := tx.MoveTakenRatioTotal(ctx, "p-1", "0.4"); err != nil {
				return err
			}
			return tx.ResolvePending(ctx, "p-1")
		},
	}); err != nil {
		t.Fatal(err)
	}

	res, err := j.RecordFill(ctx, observation("o-1", "4"))
	if err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if !res.Changed || res.Delta != "4" {
		t.Fatalf("fill result = %+v, want a delta of 4", res)
	}
	if order != "project,exit" {
		t.Errorf("hook order = %q, want the projection first so the exit sees it", order)
	}
	if projectedDelta != "4" {
		t.Errorf("the projection saw delta %q, want the new quantity 4", projectedDelta)
	}
	if got := takenRatio(t, j, "p-1"); got != "0.4" {
		t.Errorf("taken_ratio_total = %q, want 0.4", got)
	}

	var quantity, pendingAction string
	if err := j.db.QueryRowContext(ctx,
		"SELECT quantity FROM positions WHERE id = 'p-1'").Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if quantity != "4" {
		t.Errorf("projected quantity = %q, want 4", quantity)
	}
	if err := j.db.QueryRowContext(ctx,
		"SELECT coalesce(pending_action,'') FROM exit_states WHERE position_id = 'p-1'").
		Scan(&pendingAction); err != nil {
		t.Fatal(err)
	}
	if pendingAction != "" {
		t.Errorf("pending_action = %q, want the proposal resolved by the fill it answered", pendingAction)
	}
}

// TestAFailingApplyHookRollsBackTheFill is the atomicity claim from the other
// side: the projection could not be applied, so the fill did not happen either.
// Anything less would leave a fill the projection never saw, and the next
// observation would measure its delta from a quantity nothing believes.
func TestAFailingApplyHookRollsBackTheFill(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()

	boom := errors.New("the position state machine refused this transition")
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(ctx context.Context, tx *ApplyTx, _ AppliedFill) error {
			// Write something first, so the rollback has something to undo.
			if _, err := tx.Exec(ctx,
				"UPDATE positions SET quantity = '99' WHERE id = 'p-1'"); err != nil {
				return err
			}
			return boom
		},
		Exit: func(ctx context.Context, tx *ApplyTx, _ AppliedFill) error {
			t.Error("the exit hook must not run after the projection failed")
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := j.RecordFill(ctx, observation("o-1", "4"))
	if !errors.Is(err, boom) {
		t.Fatalf("RecordFill must surface the hook's error, got %v", err)
	}

	// (a) the snapshot did not advance — there is no row at all, because this was
	// the order's first observation.
	if _, err := j.LookupFill(ctx, "o-1"); !errors.Is(err, ErrFillNotFound) {
		t.Errorf("the fill snapshot survived a failed apply: %v", err)
	}
	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Errorf("fill events = %d, want none", len(events))
	}
	// (b) the hook's own write is gone with it.
	var quantity string
	if err := j.db.QueryRowContext(ctx,
		"SELECT quantity FROM positions WHERE id = 'p-1'").Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if quantity != "10" {
		t.Errorf("position quantity = %q, want the pre-fill 10", quantity)
	}
	// (c) nothing is lost: the order is still tracked, so the next poll offers the
	// same cumulative quantity again and the apply is retried rather than skipped.
	tracked, err := j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 0 {
		t.Errorf("tracked orders = %+v; a first observation that rolled back leaves no "+
			"snapshot row, and the detector re-offers the order from its open-order read", tracked)
	}
}

// TestAFailingExitHookRollsBackTheProjectionToo: the second hook's failure must
// undo the first hook's writes as well. They are one apply, not two.
func TestAFailingExitHookRollsBackTheProjectionToo(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()

	boom := errors.New("the exit policy refused this application")
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(ctx context.Context, tx *ApplyTx, _ AppliedFill) error {
			_, err := tx.Exec(ctx, "UPDATE positions SET quantity = '4' WHERE id = 'p-1'")
			return err
		},
		Exit: func(ctx context.Context, tx *ApplyTx, _ AppliedFill) error {
			if err := tx.MoveTakenRatioTotal(ctx, "p-1", "0.4"); err != nil {
				return err
			}
			return boom
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := j.RecordFill(ctx, observation("o-1", "4")); !errors.Is(err, boom) {
		t.Fatalf("want the exit hook's error, got %v", err)
	}
	var quantity string
	if err := j.db.QueryRowContext(ctx,
		"SELECT quantity FROM positions WHERE id = 'p-1'").Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if quantity != "10" {
		t.Errorf("the projection survived the exit hook's failure: quantity = %q", quantity)
	}
	if got := takenRatio(t, j, "p-1"); got != "0" {
		t.Errorf("taken_ratio_total = %q, want the pre-fill 0", got)
	}
	if _, err := j.LookupFill(ctx, "o-1"); !errors.Is(err, ErrFillNotFound) {
		t.Errorf("the fill snapshot survived a failed apply: %v", err)
	}
}

// TestApplyHooksSkipRefusedAndNoOpSnapshots: nothing was applied, so there is
// nothing to apply. A hook firing on a fail-closed snapshot would be projecting
// a fill the ledger just declined to believe.
func TestApplyHooksSkipRefusedAndNoOpSnapshots(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()

	var calls int
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(context.Context, *ApplyTx, AppliedFill) error {
			calls++
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	// A caller-side fail-closed verdict.
	refused := observation("o-refused", "3")
	refused.FailClosed = true
	refused.Reason = ReasonFillStateUnknown
	if _, err := j.RecordFill(ctx, refused); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("the hook ran for a refused snapshot (%d calls)", calls)
	}

	// A real fill, then the identical snapshot again.
	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("hook calls after one real fill = %d, want 1", calls)
	}
	repeat := observation("o-1", "3")
	repeat.ObservedAt = "2026-03-30T00:30:05Z"
	if _, err := j.RecordFill(ctx, repeat); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("hook calls after a byte-identical re-observation = %d, want it unchanged at 1", calls)
	}

	// A decrease fails closed inside the ledger itself.
	if _, err := j.RecordFill(ctx, observation("o-1", "1")); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("the hook ran for a shrinking cumulative quantity (%d calls)", calls)
	}
}

// TestUnboundHooksLeaveRecordFillUnchanged is the regression: a journal with no
// hooks — every journal until the domain is wired — records fills exactly as it
// did before this change.
func TestUnboundHooksLeaveRecordFillUnchanged(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	res, err := j.RecordFill(ctx, observation("o-1", "3"))
	if err != nil {
		t.Fatalf("RecordFill with no hooks: %v", err)
	}
	if !res.Changed || res.Delta != "3" {
		t.Fatalf("fill result = %+v, want the unchanged behaviour", res)
	}
}

// TestApplyHooksAreBoundOnce: two projections of the same fill stream is not a
// configuration, and a second binding would silently become the truth from an
// arbitrary fill onwards.
func TestApplyHooksAreBoundOnce(t *testing.T) {
	j := openTestJournal(t)
	noop := func(context.Context, *ApplyTx, AppliedFill) error { return nil }

	if err := j.SetApplyHooks(ApplyHooks{}); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("binding nothing must be refused, got %v", err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Project: noop}); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Exit: noop}); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("rebinding must be refused, got %v", err)
	}
}

// TestASmuggledApplyHandleIsDead is rule 5: the handle is valid for the call it
// was given to. A hook that stores it and uses it later is writing into a
// transaction that has already ended, and that must be an error rather than a
// silent write into whatever comes next.
func TestASmuggledApplyHandleIsDead(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()

	var smuggled *ApplyTx
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(_ context.Context, tx *ApplyTx, _ AppliedFill) error {
			smuggled = tx
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}
	if smuggled == nil {
		t.Fatal("the hook did not run")
	}

	if err := smuggled.MoveTakenRatioTotal(ctx, "p-1", "0.4"); !errors.Is(err, ErrApplyTxClosed) {
		t.Errorf("a stored handle must refuse to move the taken ratio, got %v", err)
	}
	if err := smuggled.ResolvePending(ctx, "p-1"); !errors.Is(err, ErrApplyTxClosed) {
		t.Errorf("a stored handle must refuse to resolve a proposal, got %v", err)
	}
	if _, err := smuggled.Exec(ctx, "UPDATE positions SET quantity = '99' WHERE id = 'p-1'"); !errors.Is(err, ErrApplyTxClosed) {
		t.Errorf("a stored handle must refuse to write, got %v", err)
	}
	if _, err := smuggled.Query(ctx, "SELECT 1"); !errors.Is(err, ErrApplyTxClosed) {
		t.Errorf("a stored handle must refuse to read, got %v", err)
	}
	if _, err := smuggled.PendingState(ctx, "p-1"); !errors.Is(err, ErrApplyTxClosed) {
		t.Errorf("a stored handle must refuse to read the exit state, got %v", err)
	}
	if got := takenRatio(t, j, "p-1"); got != "0" {
		t.Errorf("taken_ratio_total = %q; the dead handle wrote something", got)
	}
}

// TestApplyTxExecRefusesTheGuardedColumns: the general Exec exists for the
// projection tables, whose shape the state machine owns. Reaching the guarded
// exit columns through it would put a write of those columns in a file the
// layout test cannot see, so the statement is refused.
func TestApplyTxExecRefusesTheGuardedColumns(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()

	var refusals []error
	if err := j.SetApplyHooks(ApplyHooks{
		Exit: func(ctx context.Context, tx *ApplyTx, _ AppliedFill) error {
			for _, stmt := range []string{
				"UPDATE exit_states SET taken_ratio_total = '1' WHERE position_id = 'p-1'",
				"UPDATE exit_states SET pending_action = NULL WHERE position_id = 'p-1'",
				"UPDATE exit_states SET pending_level = NULL WHERE position_id = 'p-1'",
				"UPDATE exit_states SET pending_intent_id = NULL WHERE position_id = 'p-1'",
			} {
				_, err := tx.Exec(ctx, stmt)
				refusals = append(refusals, err)
			}
			// The projection tables are not guarded: that is what Exec is for.
			_, err := tx.Exec(ctx, "UPDATE positions SET quantity = '3' WHERE id = 'p-1'")
			return err
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if len(refusals) != 4 {
		t.Fatalf("guarded statements attempted = %d, want 4", len(refusals))
	}
	for i, err := range refusals {
		if !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("guarded statement %d was not refused: %v", i, err)
		}
	}
	if got := takenRatio(t, j, "p-1"); got != "0" {
		t.Errorf("taken_ratio_total = %q, want it untouched by the refused statements", got)
	}
}

// TestTakenRatioIsMonotoneAndBounded: the journal does not own the denominator
// rule, but it does own the two things it can check without owning it. A
// decrease would mean the position un-took a take-profit, which is not an event
// that exists; a value above 1 would mean more than the whole position was
// taken.
func TestTakenRatioIsMonotoneAndBounded(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()

	var errs []error
	if err := j.SetApplyHooks(ApplyHooks{
		Exit: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
			switch fill.Delta {
			case "3":
				return tx.MoveTakenRatioTotal(ctx, "p-1", "0.4")
			default:
				errs = append(errs,
					tx.MoveTakenRatioTotal(ctx, "p-1", "0.2"),  // backwards
					tx.MoveTakenRatioTotal(ctx, "p-1", "1.5"),  // more than the position
					tx.MoveTakenRatioTotal(ctx, "p-1", "-0.1"), // negative
					tx.MoveTakenRatioTotal(ctx, "p-1", "many"), // not a decimal
				)
				// A non-decreasing move is fine, including staying put.
				if err := tx.MoveTakenRatioTotal(ctx, "p-1", "0.4"); err != nil {
					return err
				}
				return tx.MoveTakenRatioTotal(ctx, "p-1", "0.65")
			}
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}
	if got := takenRatio(t, j, "p-1"); got != "0.4" {
		t.Fatalf("taken_ratio_total = %q, want 0.4", got)
	}
	if _, err := j.RecordFill(ctx, observation("o-1", "7")); err != nil {
		t.Fatal(err)
	}
	for i, err := range errs {
		if !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("refusal %d: %v, want ErrInvalidRequest", i, err)
		}
	}
	if got := takenRatio(t, j, "p-1"); got != "0.65" {
		t.Errorf("taken_ratio_total = %q, want the accepted 0.65", got)
	}
}

// TestPendingStateReadsWhatTheApplyPointOwns covers the read side and the
// absent-row case: an externally acquired position has no exit state, and that
// is a fact for the hook to handle rather than an error for the journal to
// invent a row over.
func TestPendingStateReadsWhatTheApplyPointOwns(t *testing.T) {
	j := applyHookFixture(t)
	ctx := context.Background()
	insertPosition2(t, j, "p-external", nil, 2)
	armPendingForTest(t, j, "p-1")

	var (
		pending PendingState
		absent  error
	)
	if err := j.SetApplyHooks(ApplyHooks{
		Exit: func(ctx context.Context, tx *ApplyTx, _ AppliedFill) error {
			var err error
			if pending, err = tx.PendingState(ctx, "p-1"); err != nil {
				return err
			}
			_, absent = tx.PendingState(ctx, "p-external")
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}

	if !pending.Pending() || pending.PendingAction != "PARTIAL_TAKE" ||
		pending.PendingLevel != RatchetPartialLock || pending.PendingIntentID != "intent-1" {
		t.Errorf("pending state = %+v, want the armed proposal", pending)
	}
	if pending.TakenRatioTotal != "0" || pending.Completed {
		t.Errorf("pending state = %+v, want nothing taken and not completed", pending)
	}
	if !errors.Is(absent, ErrExitStateNotFound) {
		t.Errorf("a position with no exit state: %v, want ErrExitStateNotFound", absent)
	}
}

// --- layout: the columns have exactly one writer ----------------------------

// TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook is the structural half of
// "taken_ratio·pending은 apply hook에서만". A behavioural test can only show
// that today's code path does not write them elsewhere; this shows that no
// other file in the package mentions them at all, so a second writer cannot
// appear without someone editing this list and justifying it in review.
func TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook(t *testing.T) {
	t.Parallel()

	// core_domain.go declares the columns (the DDL); apply_hook.go is the only
	// code that reads or writes them. Nothing else may name them.
	allowed := map[string]bool{"core_domain.go": true, "apply_hook.go": true}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	checked, writer := 0, false
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		checked++
		text := string(source)
		for _, column := range guardedColumns {
			mentions := strings.Contains(text, column)
			switch {
			case mentions && !allowed[name]:
				t.Errorf("%s names exit_states.%s: that column moves only at the atomic apply "+
					"point (apply_hook.go), and a second writer would make "+
					"\"taken_ratio·pending 해소는 여기서만\" false", name, column)
			case mentions && name == "apply_hook.go":
				writer = true
			}
		}
	}
	// Positive controls: an empty walk, or a hook file that stopped writing the
	// columns, would make the assertion above pass for the wrong reason.
	if checked == 0 {
		t.Fatal("no production files were inspected; the walk is broken")
	}
	if !writer {
		t.Fatal("apply_hook.go does not write the guarded columns; the rule is protecting nothing")
	}
	if len(guardedColumns) != 4 {
		t.Fatalf("guardedColumns = %v, want the four fill-time exit fields", guardedColumns)
	}
}

// TestNoExportedAPIHandsOutAnApplyTx is the reachability half. The guarded
// writers are methods on *ApplyTx, so "they can only be called inside a hook"
// is true exactly as long as nothing outside this package can obtain one: no
// exported function returns it, and every field is unexported, so the zero
// value is the only one a caller can build and its methods refuse.
func TestNoExportedAPIHandsOutAnApplyTx(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}
	pkg, ok := pkgs["journal"]
	if !ok {
		t.Fatalf("package journal not found in %v", pkgs)
	}

	inspected, sawStruct := 0, false
	for name, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() || d.Type.Results == nil {
					continue
				}
				inspected++
				for _, result := range d.Type.Results.List {
					if isApplyTxPointer(result.Type) {
						t.Errorf("%s: exported %s returns *ApplyTx; the handle must be reachable "+
							"only from inside a hook call", name, d.Name.Name)
					}
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name.Name != "ApplyTx" {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						t.Fatalf("ApplyTx is not a struct")
					}
					sawStruct = true
					for _, field := range st.Fields.List {
						for _, fieldName := range field.Names {
							if fieldName.IsExported() {
								t.Errorf("ApplyTx.%s is exported: a caller could then build a "+
									"working handle without going through a hook", fieldName.Name)
							}
						}
					}
				}
			}
		}
	}
	if inspected == 0 {
		t.Fatal("no exported functions were inspected; the walk is broken")
	}
	if !sawStruct {
		t.Fatal("the ApplyTx declaration was not found; the walk is broken")
	}
}

func isApplyTxPointer(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "ApplyTx"
}
