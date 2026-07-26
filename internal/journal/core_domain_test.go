package journal

import (
	"context"
	"strings"
	"testing"
)

// core_domain_test.go is the schema contract for journal v6 (change
// add-core-domain, task 0.1). Every assertion here corresponds to a line of
// design.md D7: the migration is a transcription of that table, so the tests
// read it back column by column and constraint by constraint.
//
// The constraints are exercised behaviourally — insert a row that violates one
// and require a refusal — rather than by string-matching the DDL, because what
// protects a live account is the refusal and not the text. The two partial
// indexes are additionally checked in sqlite_master: their WHERE clause is the
// point of them, and a behavioural test cannot tell a missing predicate from a
// test that never wrote the excluded rows.

// insertPosition writes a minimal engine-owned position.
func insertPosition(t *testing.T, j *Journal, id string, decisionID any) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		    quantity, avg_price, opened_at)
		 VALUES (?, 'acct-1', 'kr', '005930', 1, ?, ?, '10', '70000', '2026-03-30T00:30:00Z')`,
		id, decisionID, PositionOpen); err != nil {
		t.Fatalf("insert position %s: %v", id, err)
	}
}

// insertExitState writes a minimal RATCHET exit state for a position.
func insertExitState(t *testing.T, j *Journal, positionID string) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO exit_states
		   (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		    baseline_price, high_water, ratchet_level, updated_at)
		 VALUES (?, ?, '70000', '68000', '2000', '68000', '70000', ?, '2026-03-30T00:30:00Z')`,
		positionID, ExitPolicyRatchet, RatchetNone); err != nil {
		t.Fatalf("insert exit state for %s: %v", positionID, err)
	}
}

// TestPositionsStateCheck pins the state enum. The transition table itself is
// task 6.1; what the schema guarantees is that no value outside the six can be
// stored at all, so an unhandled state cannot arrive from the database.
func TestPositionsStateCheck(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	for i, state := range []string{
		PositionFlat, PositionOpening, PositionOpen,
		PositionScaling, PositionClosing, PositionClosed,
	} {
		if _, err := j.db.ExecContext(ctx,
			`INSERT INTO positions
			   (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
			 VALUES (?, 'acct-1', 'kr', '005930', ?, ?, '10', '70000')`,
			"p-"+state, i+1, state); err != nil {
			t.Fatalf("state %s must be accepted: %v", state, err)
		}
	}

	_, err := j.db.ExecContext(ctx,
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		 VALUES ('p-bad', 'acct-1', 'kr', '005930', 99, 'HALF_OPEN', '10', '70000')`)
	if err == nil {
		t.Fatal("an unknown position state must violate the CHECK constraint")
	}
	if !strings.Contains(strings.ToUpper(err.Error()), "CHECK") {
		t.Fatalf("want a CHECK constraint failure, got %v", err)
	}

	// The CHECK passes on NULL in SQLite, so the column carries NOT NULL as well:
	// a stateless position would be a position no state machine can advance.
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		 VALUES ('p-null', 'acct-1', 'kr', '005930', 98, NULL, '10', '70000')`); err == nil {
		t.Fatal("a NULL position state must be refused")
	}
}

// TestPositionsEntryDecisionReference is D7's nullable foreign key: an engine
// position names the decision that justified it, and an externally acquired one
// names nothing — but it may never name a decision that does not exist.
func TestPositionsEntryDecisionReference(t *testing.T) {
	j := openTestJournal(t)
	insertDecision(t, j, "d-1", "nonce-1")

	insertPosition(t, j, "p-engine", "d-1")
	// External/manual positions carry NULL: there is no decision, and inventing
	// one would make them exit-policy targets with no baseline to protect.
	insertPosition2(t, j, "p-external", nil, 2)

	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		    quantity, avg_price)
		 VALUES ('p-ghost','acct-1','kr','005930',3,'d-missing',?, '10','70000')`,
		PositionOpen); err == nil {
		t.Fatal("a position naming an unrecorded decision must violate the foreign key")
	}
}

// insertPosition2 is insertPosition with a caller-chosen instance, for the tests
// that need more than one instance of the same symbol.
func insertPosition2(t *testing.T, j *Journal, id string, decisionID any, seq int) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		    quantity, avg_price, opened_at)
		 VALUES (?, 'acct-1', 'kr', '005930', ?, ?, ?, '10', '70000', '2026-03-30T00:30:00Z')`,
		id, seq, decisionID, PositionOpen); err != nil {
		t.Fatalf("insert position %s: %v", id, err)
	}
}

// TestPositionsInstanceUnique pins the re-entry rule: CLOSED is final, and the
// next entry on the same symbol is a new instance. Two rows claiming the same
// instance would let a re-entry silently merge into the position it re-entered.
func TestPositionsInstanceUnique(t *testing.T) {
	j := openTestJournal(t)
	insertPosition(t, j, "p-1", nil)

	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		 VALUES ('p-2','acct-1','kr','005930',1,?, '10','70000')`, PositionOpen); err == nil {
		t.Fatal("two positions cannot share one (account, market, symbol, instance)")
	}
	// A new instance of the same symbol is exactly what a re-entry is.
	insertPosition2(t, j, "p-2", nil, 2)
	// So is the same symbol on another market.
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		 VALUES ('p-3','acct-1','us','005930',1,?, '10','70000')`, PositionOpen); err != nil {
		t.Fatalf("the same symbol on another market is another position: %v", err)
	}
}

// TestPositionAdjustmentsConstraints covers the adjustment row's enum, its
// foreign key and the compare half of compare-and-append: an adjustment with no
// expected previous quantity cannot be compared against anything, so the column
// is NOT NULL rather than a convenience.
func TestPositionAdjustmentsConstraints(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "p-1", nil)

	insert := func(id, positionID, kind string, expectedPrev any) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO position_adjustments
			   (id, position_id, kind, expected_prev_quantity, prev_quantity, new_quantity,
			    broker_as_of, evidence, created_at)
			 VALUES (?,?,?,?, '10', '8', '2026-03-30T00:29:55Z', 'holdings read',
			         '2026-03-30T00:30:00Z')`,
			id, positionID, kind, expectedPrev)
		return err
	}

	for _, kind := range []string{AdjustmentExternal, AdjustmentManual, AdjustmentUnknown} {
		if err := insert("adj-"+kind, "p-1", kind, "10"); err != nil {
			t.Fatalf("kind %s must be accepted: %v", kind, err)
		}
	}
	if err := insert("adj-bad-kind", "p-1", "SPLIT", "10"); err == nil {
		t.Error("an unknown adjustment kind must violate the CHECK constraint")
	}
	if err := insert("adj-orphan", "p-missing", AdjustmentExternal, "10"); err == nil {
		t.Error("an adjustment without its position must violate the foreign key")
	}
	if err := insert("adj-no-expected", "p-1", AdjustmentExternal, nil); err == nil {
		t.Error("an adjustment with no expected previous quantity cannot be compared; it must be refused")
	}

	// The average prices default to "not observed" rather than to NULL, the same
	// convention fill_snapshots.average_price uses.
	var prevAvg, newAvg string
	if err := j.db.QueryRowContext(ctx,
		`SELECT prev_avg_price, new_avg_price FROM position_adjustments WHERE id = ?`,
		"adj-"+AdjustmentExternal).Scan(&prevAvg, &newAvg); err != nil {
		t.Fatal(err)
	}
	if prevAvg != "" || newAvg != "" {
		t.Errorf("adjustment average prices = %q/%q, want the empty string", prevAvg, newAvg)
	}
}

// TestOperatingModesConstraints pins D3's three modes and the actor that
// distinguishes an automatic tightening from an operator's decision.
func TestOperatingModesConstraints(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	insert := func(id, mode, actor string, cause any) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO operating_modes (id, account_ref, mode, cause, actor, created_at)
			 VALUES (?, 'acct-1', ?, ?, ?, '2026-03-30T00:30:00Z')`, id, mode, cause, actor)
		return err
	}

	for _, mode := range []string{ModeNormal, ModeEntryBlocked, ModeHaltAll} {
		if err := insert("m-"+mode, mode, ModeActorAuto, "daily loss limit reached"); err != nil {
			t.Fatalf("mode %s must be accepted: %v", mode, err)
		}
	}
	// EXIT_ONLY was deliberately dropped (D3): it behaved identically to
	// ENTRY_BLOCKED, and the schema must not let it back in by accident.
	if err := insert("m-exit-only", "EXIT_ONLY", ModeActorAuto, "cause"); err == nil {
		t.Error("EXIT_ONLY was removed from the mode set and must violate the CHECK constraint")
	}
	if err := insert("m-bad-actor", ModeNormal, "SCHEDULER", "cause"); err == nil {
		t.Error("an unknown actor must violate the CHECK constraint")
	}
	if err := insert("m-no-cause", ModeEntryBlocked, ModeActorOperator, nil); err == nil {
		t.Error("a mode transition with no cause is not auditable and must be refused")
	}
	if err := insert("m-operator", ModeNormal, ModeActorOperator, "approved by operator"); err != nil {
		t.Fatalf("an operator transition must be accepted: %v", err)
	}
}

// TestExitStatesRequiredColumns is D5's first correction stated as a
// constraint: baseline_price, high_water and the three entry values are NOT
// NULL with no default, so there is no representable exit state that has an
// open position and no protection level.
func TestExitStatesRequiredColumns(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "p-1", nil)

	const insert = `INSERT INTO exit_states
		  (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		   baseline_price, high_water, ratchet_level, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`

	base := []any{"p-1", ExitPolicyRatchet, "70000", "68000", "2000", "68000", "70000",
		RatchetNone, "2026-03-30T00:30:00Z"}
	for i, column := range []string{
		"position_id", "policy_kind", "entry_price", "initial_stop", "initial_risk",
		"baseline_price", "high_water", "ratchet_level", "updated_at",
	} {
		args := append([]any(nil), base...)
		args[i] = nil
		if _, err := j.db.ExecContext(ctx, insert, args...); err == nil {
			t.Errorf("a NULL %s must be refused", column)
			// Undo, so the next case does not collide on the primary key.
			if _, derr := j.db.ExecContext(ctx,
				"DELETE FROM exit_states WHERE position_id = 'p-1'"); derr != nil {
				t.Fatal(derr)
			}
		}
	}

	if _, err := j.db.ExecContext(ctx, insert, base...); err != nil {
		t.Fatalf("a complete exit state must be accepted: %v", err)
	}
	// One exit state per position: the primary key is the position.
	if _, err := j.db.ExecContext(ctx, insert, base...); err == nil {
		t.Error("a position has one exit state, not two")
	}
}

// TestExitStatesEnumsAndDefaults pins the policy enum (one policy per position),
// the ratchet level enum, and the two columns D7 gives defaults to.
func TestExitStatesEnumsAndDefaults(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	for i, kind := range []string{ExitPolicyRatchet, ExitPolicyLadder} {
		id := "p-" + kind
		insertPosition2(t, j, id, nil, i+1)
		if _, err := j.db.ExecContext(ctx,
			`INSERT INTO exit_states
			   (position_id, policy_kind, entry_price, initial_stop, initial_risk,
			    baseline_price, high_water, ratchet_level, updated_at)
			 VALUES (?,?, '70000','68000','2000','68000','70000', ?, '2026-03-30T00:30:00Z')`,
			id, kind, RatchetNone); err != nil {
			t.Fatalf("policy kind %s must be accepted: %v", kind, err)
		}
	}
	insertPosition2(t, j, "p-hybrid", nil, 3)
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO exit_states
		   (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		    baseline_price, high_water, ratchet_level, updated_at)
		 VALUES ('p-hybrid','BOTH','70000','68000','2000','68000','70000',?,'2026-03-30T00:30:00Z')`,
		RatchetNone); err == nil {
		t.Error("a position has one policy: an unknown policy_kind must violate the CHECK constraint")
	}

	for _, level := range []string{
		RatchetNone, RatchetHalfRisk, RatchetBreakeven, RatchetPartialLock, RatchetProfitLock,
	} {
		if _, err := j.db.ExecContext(ctx,
			"UPDATE exit_states SET ratchet_level = ? WHERE position_id = 'p-RATCHET'",
			level); err != nil {
			t.Fatalf("ratchet level %s must be accepted: %v", level, err)
		}
	}
	if _, err := j.db.ExecContext(ctx,
		"UPDATE exit_states SET ratchet_level = 'FULL_LOCK' WHERE position_id = 'p-RATCHET'"); err == nil {
		t.Error("an unknown ratchet level must violate the CHECK constraint")
	}

	// taken_ratio_total starts at "0" and completed at 0: a fresh exit state has
	// taken nothing and is not finished.
	var taken string
	var completed int
	if err := j.db.QueryRowContext(ctx,
		`SELECT taken_ratio_total, completed FROM exit_states WHERE position_id = 'p-RATCHET'`).
		Scan(&taken, &completed); err != nil {
		t.Fatal(err)
	}
	if taken != "0" || completed != 0 {
		t.Errorf("defaults = taken %q completed %d, want \"0\" and 0", taken, completed)
	}
	// Both are NOT NULL, so the default cannot be defeated by writing NULL over it.
	if _, err := j.db.ExecContext(ctx,
		"UPDATE exit_states SET taken_ratio_total = NULL WHERE position_id = 'p-RATCHET'"); err == nil {
		t.Error("taken_ratio_total is NOT NULL")
	}
	if _, err := j.db.ExecContext(ctx,
		"UPDATE exit_states SET completed = NULL WHERE position_id = 'p-RATCHET'"); err == nil {
		t.Error("completed is NOT NULL")
	}
}

// TestExitStatesPendingColumnsAreFree pins the pending triple's shape: three
// nullable columns, no foreign key on the intent id, and pending_level carrying
// either a ratchet level or a ladder rung index.
//
// The absent foreign key is deliberate. A proposal is armed *before* the intent
// exists — that ordering is what makes a crash between arming and submitting
// recoverable — so a reference constraint here would forbid the only sequence
// the requirement allows.
func TestExitStatesPendingColumnsAreFree(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "p-1", nil)
	insertExitState(t, j, "p-1")

	for _, level := range []string{RatchetPartialLock, "2"} {
		if _, err := j.db.ExecContext(ctx,
			`UPDATE exit_states
			    SET pending_action = 'PARTIAL_TAKE', pending_level = ?, pending_intent_id = 'intent-not-yet'
			  WHERE position_id = 'p-1'`, level); err != nil {
			t.Fatalf("pending_level %q must be storable: %v", level, err)
		}
	}
	// Resolving clears all three together.
	if _, err := j.db.ExecContext(ctx,
		`UPDATE exit_states
		    SET pending_action = NULL, pending_level = NULL, pending_intent_id = NULL
		  WHERE position_id = 'p-1'`); err != nil {
		t.Fatalf("clearing the pending triple: %v", err)
	}
}

// TestExitStatesRequireAPosition keeps orphaned protection out: an exit state is
// the protection *of* a position, and one without a position protects nothing.
func TestExitStatesRequireAPosition(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO exit_states
		   (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		    baseline_price, high_water, ratchet_level, updated_at)
		 VALUES ('p-missing',?, '70000','68000','2000','68000','70000',?,'2026-03-30T00:30:00Z')`,
		ExitPolicyRatchet, RatchetNone); err == nil {
		t.Fatal("an exit state without its position must violate the foreign key")
	}
}

// TestExitEventsAreAppendOnlyHistory pins the judgement log: an autoincrementing
// id (the order is the record), a required position and a required timestamp,
// and every judgement column nullable because a no-op observation is still a
// judgement worth having in the provenance chain.
func TestExitEventsAreAppendOnlyHistory(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "p-1", nil)

	for i := 0; i < 2; i++ {
		if _, err := j.db.ExecContext(ctx,
			`INSERT INTO exit_events
			   (position_id, observed_price, high_water, baseline_after, level_after,
			    action, proposed_intent_id, created_at)
			 VALUES ('p-1','70500','70500','68000',?, 'HOLD', NULL, '2026-03-30T00:30:0'||?||'Z')`,
			RatchetNone, i); err != nil {
			t.Fatalf("appending an exit event: %v", err)
		}
	}
	// A judgement that observed nothing and proposed nothing is still a row.
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO exit_events (position_id, created_at)
		 VALUES ('p-1','2026-03-30T00:30:05Z')`); err != nil {
		t.Fatalf("a judgement with no price must still be recordable: %v", err)
	}

	var ids []int64
	rows, err := j.db.QueryContext(ctx, "SELECT id FROM exit_events ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[0] >= ids[1] || ids[1] >= ids[2] {
		t.Errorf("exit event ids = %v, want three strictly increasing", ids)
	}

	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO exit_events (position_id, created_at)
		 VALUES ('p-missing','2026-03-30T00:30:05Z')`); err == nil {
		t.Error("an exit event without its position must violate the foreign key")
	}
	if _, err := j.db.ExecContext(ctx,
		"INSERT INTO exit_events (position_id, created_at) VALUES ('p-1', NULL)"); err == nil {
		t.Error("an exit event with no timestamp has no place in an ordered history")
	}
}

// TestTradeOutcomesFrozenRecord pins the closed-position record: one row per
// position, the four decimals that make realised R reconstructible, and the
// policy-specific exit columns that are nullable because a position runs one
// policy and therefore has only one of them.
func TestTradeOutcomesFrozenRecord(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "p-1", nil)

	const insert = `INSERT INTO trade_outcomes
		  (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
		VALUES (?,?,?,?,?,?,?,?,?)`

	base := []any{"p-1", "120000", "1.5", "2000", "10", 3600, RatchetProfitLock, nil,
		"2026-03-30T01:30:00Z"}
	for i, column := range []string{
		"position_id", "realized_pnl_after_costs", "realized_r", "initial_risk",
		"initial_quantity",
	} {
		args := append([]any(nil), base...)
		args[i] = nil
		if _, err := j.db.ExecContext(ctx, insert, args...); err == nil {
			t.Errorf("a NULL %s must be refused: the frozen record would not reconstruct", column)
			if _, derr := j.db.ExecContext(ctx,
				"DELETE FROM trade_outcomes WHERE position_id = 'p-1'"); derr != nil {
				t.Fatal(derr)
			}
		}
	}
	closedAt := append([]any(nil), base...)
	closedAt[8] = nil
	if _, err := j.db.ExecContext(ctx, insert, closedAt...); err == nil {
		t.Error("a NULL closed_at must be refused: the retention sweep has nothing to age")
	}

	if _, err := j.db.ExecContext(ctx, insert, base...); err != nil {
		t.Fatalf("a complete outcome must be accepted: %v", err)
	}
	if _, err := j.db.ExecContext(ctx, insert, base...); err == nil {
		t.Error("a position closes once: the outcome is keyed by the position")
	}

	// A LADDER position carries a rung and no ratchet level; the columns are
	// nullable exactly so neither policy has to invent the other's field.
	insertPosition2(t, j, "p-2", nil, 2)
	ladder := append([]any(nil), base...)
	ladder[0], ladder[6], ladder[7] = "p-2", nil, 3
	if _, err := j.db.ExecContext(ctx, insert, ladder...); err != nil {
		t.Fatalf("a ladder outcome must be accepted: %v", err)
	}

	if _, err := j.db.ExecContext(ctx, insert,
		"p-missing", "1", "1", "1", "1", 1, nil, nil, "2026-03-30T01:30:00Z"); err == nil {
		t.Error("an outcome without its position must violate the foreign key")
	}
}

// TestCoreDomainPartialIndexPredicates reads the v6 partial indexes back out of
// sqlite_master. The WHERE clause is what makes them the crash-recovery scan and
// the observation loop's working set rather than a full-table index, and a
// behavioural test cannot see a missing predicate at all.
func TestCoreDomainPartialIndexPredicates(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	want := map[string][]string{
		"idx_exit_states_pending": {
			"exit_states", "pending_intent_id", "WHERE pending_action IS NOT NULL",
		},
		"idx_exit_states_open": {
			"exit_states", "completed", "WHERE completed = 0",
		},
	}
	for name, fragments := range want {
		var ddl string
		if err := j.db.QueryRowContext(ctx,
			"SELECT sql FROM sqlite_master WHERE type='index' AND name = ?", name).Scan(&ddl); err != nil {
			t.Errorf("index %s: %v", name, err)
			continue
		}
		for _, fragment := range fragments {
			if !strings.Contains(ddl, fragment) {
				t.Errorf("index %s is missing %q:\n%s", name, fragment, ddl)
			}
		}
	}
}

// TestCoreDomainTablesAreStrict keeps the v6 tables inside the convention the
// rest of the journal holds to: STRICT, so a decimal string cannot be silently
// stored as a float and an unconvertible value is refused rather than coerced.
func TestCoreDomainTablesAreStrict(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	for _, table := range []string{
		"positions", "position_adjustments", "operating_modes",
		"exit_states", "exit_events", "trade_outcomes",
	} {
		var ddl string
		if err := j.db.QueryRowContext(ctx,
			"SELECT sql FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&ddl); err != nil {
			t.Errorf("table %s: %v", table, err)
			continue
		}
		if !strings.Contains(strings.ToUpper(ddl), "STRICT") {
			t.Errorf("table %s is not STRICT:\n%s", table, ddl)
		}
	}

	insertPosition(t, j, "p-1", nil)
	// STRICT refuses text in an INTEGER column: instance_seq and held_seconds are
	// counts, not decimal strings, and the difference must not be papered over.
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO positions (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		 VALUES ('p-2','acct-1','kr','005930','two',?, '10','70000')`, PositionOpen); err == nil {
		t.Error("a text instance_seq must be refused by the STRICT table")
	}
}
