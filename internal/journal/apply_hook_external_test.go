package journal_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// apply_hook_external_test.go states the reachability rule from where it
// matters: outside the package.
//
// The in-package tests can reach unexported fields, so they can always build a
// working handle; that is exactly what a consumer cannot do, and this file is
// the proof. It compiles against the exported API only, and that API offers no
// way to obtain an *ApplyTx — the best a caller can do is construct the zero
// value, whose methods refuse.

// TestGuardedExitColumnsAreUnreachableFromOutsideAHook is task 0.3's "hook
// 밖에서 taken_ratio·pending을 쓸 수 없음", written as a consumer would hit it.
func TestGuardedExitColumnsAreUnreachableFromOutsideAHook(t *testing.T) {
	ctx := context.Background()

	// The only *ApplyTx a caller outside the package can produce. There is no
	// exported constructor, no exported field to fill in, and no exported
	// function anywhere in the package that returns one — which is what
	// TestNoExportedAPIHandsOutAnApplyTx keeps true as the package grows.
	var forged journal.ApplyTx

	if err := forged.MoveTakenRatioTotal(ctx, "p-1", "0.4"); !errors.Is(err, journal.ErrApplyTxClosed) {
		t.Errorf("MoveTakenRatioTotal on a forged handle = %v, want ErrApplyTxClosed", err)
	}
	if err := forged.ResolvePending(ctx, "p-1"); !errors.Is(err, journal.ErrApplyTxClosed) {
		t.Errorf("ResolvePending on a forged handle = %v, want ErrApplyTxClosed", err)
	}
	if _, err := forged.PendingState(ctx, "p-1"); !errors.Is(err, journal.ErrApplyTxClosed) {
		t.Errorf("PendingState on a forged handle = %v, want ErrApplyTxClosed", err)
	}
	if _, err := forged.Exec(ctx,
		"UPDATE exit_states SET taken_ratio_total = '1'"); !errors.Is(err, journal.ErrApplyTxClosed) {
		t.Errorf("Exec on a forged handle = %v, want ErrApplyTxClosed", err)
	}
	if _, err := forged.Query(ctx, "SELECT 1"); !errors.Is(err, journal.ErrApplyTxClosed) {
		t.Errorf("Query on a forged handle = %v, want ErrApplyTxClosed", err)
	}
}

// TestTheOnlyWayInIsAHook is the positive half: the same methods work, and only
// work, on the handle a hook is handed. Without it the test above would pass on
// an API that never works at all.
func TestTheOnlyWayInIsAHook(t *testing.T) {
	ctx := context.Background()
	j := openExternalTestJournal(t)

	var seeded, moved error
	if err := j.SetApplyHooks(journal.ApplyHooks{
		// The projection tables are the hook's own to write — that is what the
		// general Exec is for, and none of these statements names a guarded column.
		Project: func(ctx context.Context, tx *journal.ApplyTx, _ journal.AppliedFill) error {
			if _, seeded = tx.Exec(ctx,
				`INSERT INTO positions
				   (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
				 VALUES ('p-1','acct-1','kr','005930',1,?, '10','70000')`,
				journal.PositionOpen); seeded != nil {
				return seeded
			}
			_, seeded = tx.Exec(ctx,
				`INSERT INTO exit_states
				   (position_id, policy_kind, entry_price, initial_stop, initial_risk,
				    baseline_price, high_water, ratchet_level, updated_at)
				 VALUES ('p-1',?, '70000','68000','2000','68000','70000',?,?)`,
				journal.ExitPolicyRatchet, journal.RatchetNone, tx.Now())
			return seeded
		},
		Exit: func(ctx context.Context, tx *journal.ApplyTx, _ journal.AppliedFill) error {
			moved = tx.MoveTakenRatioTotal(ctx, "p-1", "0.4")
			if moved != nil {
				return moved
			}
			state, err := tx.PendingState(ctx, "p-1")
			if err != nil {
				return err
			}
			if state.TakenRatioTotal != "0.4" {
				return errors.New("the move was not visible inside the transaction")
			}
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-1", Symbol: "005930", Market: "kr", State: "OPEN_PARTIALLY_FILLED",
		Quantity: "10", FilledQuantity: "4", ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if seeded != nil {
		t.Fatalf("the hook could not write the projection: %v", seeded)
	}
	if moved != nil {
		t.Fatalf("the hook's handle must move the taken ratio: %v", moved)
	}
}

func openExternalTestJournal(t *testing.T) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := j.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return j
}
