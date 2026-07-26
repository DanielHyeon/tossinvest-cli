package journal

import (
	"context"
	"strings"
	"testing"
)

// execution_contract_test.go is the schema contract for journal v5 (change
// extend-execution-contract, task 0.1). Every assertion here corresponds to a
// line of design.md D9: the migration is a transcription of that table, so the
// tests read it back column by column and constraint by constraint.
//
// The constraints are exercised behaviourally (insert a row that violates them
// and require a refusal) rather than by string-matching the DDL, because what
// protects a live account is the refusal, not the text. The two partial UNIQUE
// indexes are additionally checked in sqlite_master: their WHERE clause is the
// whole point of them and an index without it would still pass an insert test
// for the rows a test happens to write.

const decisionInsertSQL = `INSERT INTO decisions
	  (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
	   risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`

// insertDecision writes a minimal valid EXPOSURE_RAISING decision.
func insertDecision(t *testing.T, j *Journal, id, nonce string) {
	t.Helper()
	_, err := j.db.ExecContext(context.Background(), decisionInsertSQL,
		id, "acct-1", 0, SafetyClassExposureRaising, PreimageKindRiskIntent,
		`{"symbol":"AAPL"}`, "hash-1", nil, `{"max_notional":"1000"}`, nonce,
		"2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z")
	if err != nil {
		t.Fatalf("insert decision %s: %v", id, err)
	}
}

// TestDecisionsSafetyClassCheck pins the class enum, including the value 2a
// never issues: PROTECTION_WEAKENING is reserved for 2c, and reserving it in the
// CHECK now is what makes that later change additive.
func TestDecisionsSafetyClassCheck(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	for i, class := range []string{
		SafetyClassExposureRaising,
		SafetyClassRiskReducing,
		SafetyClassProtectionWeakening,
	} {
		_, err := j.db.ExecContext(ctx, decisionInsertSQL,
			"d-ok-"+class, "acct-1", 0, class, PreimageKindRiskIntent, "{}", "h",
			nil, nil, "nonce-ok-"+class, "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z")
		if err != nil {
			t.Fatalf("class %d (%s) must be accepted: %v", i, class, err)
		}
	}

	_, err := j.db.ExecContext(ctx, decisionInsertSQL,
		"d-bad", "acct-1", 0, "EXPOSURE_LOWERING", PreimageKindRiskIntent, "{}", "h",
		nil, nil, "nonce-bad", "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z")
	if err == nil {
		t.Fatal("an unknown safety_class must violate the CHECK constraint")
	}
	if !strings.Contains(strings.ToUpper(err.Error()), "CHECK") {
		t.Fatalf("want a CHECK constraint failure, got %v", err)
	}
}

// TestDecisionsPreimageKindCheck pins the per-class preimage schema tag (D1):
// EXPOSURE_RAISING carries a RiskIntent, RISK_REDUCING a ReductionIntent, and
// nothing else is a preimage.
func TestDecisionsPreimageKindCheck(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	for _, kind := range []string{PreimageKindRiskIntent, PreimageKindReductionIntent} {
		_, err := j.db.ExecContext(ctx, decisionInsertSQL,
			"d-"+kind, "acct-1", 0, SafetyClassRiskReducing, kind, "{}", "h",
			nil, nil, "nonce-"+kind, "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z")
		if err != nil {
			t.Fatalf("preimage_kind %s must be accepted: %v", kind, err)
		}
	}

	_, err := j.db.ExecContext(ctx, decisionInsertSQL,
		"d-bad", "acct-1", 0, SafetyClassRiskReducing, "FREEFORM", "{}", "h",
		nil, nil, "nonce-bad", "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z")
	if err == nil {
		t.Fatal("an unknown preimage_kind must violate the CHECK constraint")
	}
}

// TestDecisionsRequiredColumns proves the preimage and its hash cannot be
// omitted. Gateway re-derives the hash from the stored preimage and compares
// (D1); a NULL preimage would turn that check into a no-op.
func TestDecisionsRequiredColumns(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	cases := map[string][]any{
		"risk_preimage": {"d-1", "acct-1", 0, SafetyClassRiskReducing, PreimageKindReductionIntent,
			nil, "h", nil, nil, "n-1", "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z"},
		"risk_hash": {"d-2", "acct-1", 0, SafetyClassRiskReducing, PreimageKindReductionIntent,
			"{}", nil, nil, nil, "n-2", "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z"},
		"account_ref": {"d-3", nil, 0, SafetyClassRiskReducing, PreimageKindReductionIntent,
			"{}", "h", nil, nil, "n-3", "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z"},
		"nonce": {"d-4", "acct-1", 0, SafetyClassRiskReducing, PreimageKindReductionIntent,
			"{}", "h", nil, nil, nil, "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z"},
		"issued_at": {"d-5", "acct-1", 0, SafetyClassRiskReducing, PreimageKindReductionIntent,
			"{}", "h", nil, nil, "n-5", nil, "2026-03-30T00:40:00Z"},
		"expires_at": {"d-6", "acct-1", 0, SafetyClassRiskReducing, PreimageKindReductionIntent,
			"{}", "h", nil, nil, "n-6", "2026-03-30T00:30:00Z", nil},
	}
	for column, args := range cases {
		if _, err := j.db.ExecContext(ctx, decisionInsertSQL, args...); err == nil {
			t.Errorf("a NULL %s must be refused", column)
		}
	}
}

// TestDecisionsGenerationDefaultsToZero pins D1's "2a에서는 항상 0": the column
// carries the default so a writer that does not yet reissue decisions cannot
// leave it unset.
func TestDecisionsGenerationDefaultsToZero(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO decisions
		   (id, account_ref, safety_class, preimage_kind, risk_preimage, risk_hash,
		    nonce, issued_at, expires_at)
		 VALUES ('d-1','acct-1',?,?,'{}','h','n-1','2026-03-30T00:30:00Z','2026-03-30T00:40:00Z')`,
		SafetyClassRiskReducing, PreimageKindReductionIntent); err != nil {
		t.Fatal(err)
	}
	var generation int
	if err := j.db.QueryRowContext(ctx,
		"SELECT generation FROM decisions WHERE id = 'd-1'").Scan(&generation); err != nil {
		t.Fatal(err)
	}
	if generation != 0 {
		t.Fatalf("generation = %d, want 0", generation)
	}
}

// TestDecisionsNonceUnique is the single-use guarantee the nonce exists for: two
// decisions can never claim the same one, whatever order the issuer writes them
// in.
func TestDecisionsNonceUnique(t *testing.T) {
	j := openTestJournal(t)
	insertDecision(t, j, "d-1", "nonce-shared")

	_, err := j.db.ExecContext(context.Background(), decisionInsertSQL,
		"d-2", "acct-1", 0, SafetyClassRiskReducing, PreimageKindReductionIntent, "{}", "h",
		nil, nil, "nonce-shared", "2026-03-30T00:30:00Z", "2026-03-30T00:40:00Z")
	if err == nil {
		t.Fatal("a reused nonce must violate the UNIQUE constraint")
	}
}

// TestRiskReservationsConstraints covers the reservation row's enums, its
// default state and the foreign key to the decision that justified it (D5).
func TestRiskReservationsConstraints(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertDecision(t, j, "d-1", "nonce-1")

	insert := func(id, decisionID, kind, state string) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO risk_reservations
			   (id, decision_id, account_ref, kind, amount, currency, snapshot_as_of, state)
			 VALUES (?,?, 'acct-1', ?, '100.5', 'KRW', '2026-03-30T00:29:55Z', ?)`,
			id, decisionID, kind, state)
		return err
	}

	for _, kind := range []string{ReservationKindOpenExposure, ReservationKindDailyLoss, ReservationKindCash} {
		if err := insert("r-"+kind, "d-1", kind, ReservationHeld); err != nil {
			t.Fatalf("kind %s must be accepted: %v", kind, err)
		}
	}
	if err := insert("r-bad-kind", "d-1", "MARGIN", ReservationHeld); err == nil {
		t.Error("an unknown reservation kind must violate the CHECK constraint")
	}
	if err := insert("r-bad-state", "d-1", ReservationKindCash, "PARTIAL"); err == nil {
		t.Error("an unknown reservation state must violate the CHECK constraint")
	}
	if err := insert("r-orphan", "d-missing", ReservationKindCash, ReservationHeld); err == nil {
		t.Error("a reservation without its decision must violate the foreign key")
	}

	// state defaults to HELD: a reservation is held the moment it exists.
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO risk_reservations
		   (id, decision_id, account_ref, kind, amount, currency, snapshot_as_of)
		 VALUES ('r-default','d-1','acct-1',?,'1','KRW','2026-03-30T00:29:55Z')`,
		ReservationKindCash); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := j.db.QueryRowContext(ctx,
		"SELECT state FROM risk_reservations WHERE id = 'r-default'").Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != ReservationHeld {
		t.Fatalf("default state = %q, want %q", state, ReservationHeld)
	}
}

// TestRiskReservationsReleaseReasonCheck pins the release enum. Every exit from
// a held reservation is one of four named reasons (D5); an unnamed one would
// hide an operator release inside an automatic path.
func TestRiskReservationsReleaseReasonCheck(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertDecision(t, j, "d-1", "nonce-1")

	newReservation := func(id string) {
		t.Helper()
		if _, err := j.db.ExecContext(ctx,
			`INSERT INTO risk_reservations
			   (id, decision_id, account_ref, kind, amount, currency, snapshot_as_of)
			 VALUES (?, 'd-1','acct-1',?,'1','KRW','2026-03-30T00:29:55Z')`,
			id, ReservationKindCash); err != nil {
			t.Fatal(err)
		}
	}
	release := func(id, reason string) error {
		_, err := j.db.ExecContext(ctx,
			`UPDATE risk_reservations SET state = ?, released_at = '2026-03-30T01:00:00Z',
			   release_reason = ? WHERE id = ?`, ReservationReleased, reason, id)
		return err
	}

	for _, reason := range []string{
		ReleaseReasonBrokerTerminal,
		ReleaseReasonExpiredUnconsumed,
		ReleaseReasonOperator,
		ReleaseReasonDayBoundary,
	} {
		newReservation("r-" + reason)
		if err := release("r-"+reason, reason); err != nil {
			t.Fatalf("release reason %s must be accepted: %v", reason, err)
		}
	}

	newReservation("r-bad")
	if err := release("r-bad", "BECAUSE"); err == nil {
		t.Error("an unknown release_reason must violate the CHECK constraint")
	}
}

// TestRiskReservationAttemptForeignKey pins the join D5 uses to close a
// reservation in the same transaction as the attempt's terminal record: the
// column is nullable (the reservation exists before Prepare mints the attempt)
// but a non-NULL value must name a real attempt.
func TestRiskReservationAttemptForeignKey(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertDecision(t, j, "d-1", "nonce-1")
	insertIntent(t, j, "intent-1")
	insertAttempt(t, j, "attempt-1", "intent-1")

	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO risk_reservations
		   (id, decision_id, attempt_id, account_ref, kind, amount, currency, snapshot_as_of)
		 VALUES ('r-1','d-1',NULL,'acct-1',?,'1','KRW','2026-03-30T00:29:55Z')`,
		ReservationKindOpenExposure); err != nil {
		t.Fatalf("a reservation without an attempt yet must be allowed: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		"UPDATE risk_reservations SET attempt_id = 'attempt-1' WHERE id = 'r-1'"); err != nil {
		t.Fatalf("backfilling attempt_id at Prepare time: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		"UPDATE risk_reservations SET attempt_id = 'attempt-missing' WHERE id = 'r-1'"); err == nil {
		t.Error("an unknown attempt_id must violate the foreign key")
	}
}

// TestSpentNoncesPrimaryKey proves the consumption record is the thing that
// makes a nonce single-use across restarts: a second consumption of the same
// nonce is refused by the primary key, not by a lookup the writer might skip.
func TestSpentNoncesPrimaryKey(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertDecision(t, j, "d-1", "nonce-1")

	spend := func() error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO spent_nonces (nonce, decision_id, consumed_at)
			 VALUES ('nonce-1','d-1','2026-03-30T00:31:00Z')`)
		return err
	}
	if err := spend(); err != nil {
		t.Fatalf("first consumption: %v", err)
	}
	if err := spend(); err == nil {
		t.Fatal("consuming the same nonce twice must violate the primary key")
	}

	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO spent_nonces (nonce, decision_id) VALUES ('nonce-2','d-1')`); err == nil {
		t.Error("consumed_at is NOT NULL: a consumption with no time is not a record")
	}
}

// TestReconcileActiveUnique pins the partial unique index: one active RECONCILE
// state per scope. A second row for the same scope would let a release close
// only one of them and re-open entries while the cause still holds.
func TestReconcileActiveUnique(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	enter := func(id, symbol string) error {
		var sym any
		if symbol != "" {
			sym = symbol
		}
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO reconcile_states (id, account_ref, symbol, cause, entered_at)
			 VALUES (?, 'acct-1', ?, ?, '2026-03-30T00:30:00Z')`,
			id, sym, ReconcileCauseQuantityMismatch)
		return err
	}

	if err := enter("rc-1", "AAPL"); err != nil {
		t.Fatal(err)
	}
	if err := enter("rc-2", "AAPL"); err == nil {
		t.Error("a second active state for the same (account, symbol) must be refused")
	}
	// A different symbol is a different scope.
	if err := enter("rc-3", "MSFT"); err != nil {
		t.Errorf("another symbol is a different scope: %v", err)
	}
	// Account-wide (symbol NULL) is its own scope, and it is also single: SQLite
	// treats NULLs as distinct in a UNIQUE index, so the account-wide case needs
	// its own partial index or the "one active state" rule would not hold there.
	if err := enter("rc-4", ""); err != nil {
		t.Errorf("first account-wide state: %v", err)
	}
	if err := enter("rc-5", ""); err == nil {
		t.Error("a second active account-wide state must be refused")
	}

	// Releasing frees the scope for a later entry.
	if _, err := j.db.ExecContext(ctx,
		`UPDATE reconcile_states SET released_at = '2026-03-30T01:00:00Z', release_cause = 'OPERATOR'
		 WHERE id = 'rc-1'`); err != nil {
		t.Fatal(err)
	}
	if err := enter("rc-6", "AAPL"); err != nil {
		t.Errorf("after release the scope must accept a new state: %v", err)
	}
}

// TestReconcileRequiredColumns keeps a RECONCILE row from existing without the
// facts an operator needs to act on it.
func TestReconcileRequiredColumns(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO reconcile_states (id, account_ref, cause, entered_at)
		 VALUES ('rc-1', NULL, ?, '2026-03-30T00:30:00Z')`,
		ReconcileCauseSnapshotUnavailable); err == nil {
		t.Error("account_ref is NOT NULL")
	}
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO reconcile_states (id, account_ref, cause, entered_at)
		 VALUES ('rc-2', 'acct-1', NULL, '2026-03-30T00:30:00Z')`); err == nil {
		t.Error("cause is NOT NULL")
	}
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO reconcile_states (id, account_ref, cause, entered_at)
		 VALUES ('rc-3', 'acct-1', ?, NULL)`,
		ReconcileCauseSnapshotUnavailable); err == nil {
		t.Error("entered_at is NOT NULL")
	}
}

// TestExecutionCorrectionsDedup is the crash-re-observation guarantee (D7): the
// same correction observed twice is one row, so a poll loop that restarts mid
// pass cannot double-count an amount-only correction.
func TestExecutionCorrectionsDedup(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertFillSnapshot(t, j, "order-1")

	insert := func(id, newAvg, newAmount string) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO execution_corrections
			   (id, account_ref, order_id, prev_avg_price, new_avg_price,
			    prev_filled_amount, new_filled_amount, cumulative_qty, observed_at)
			 VALUES (?, 'acct-1', 'order-1', '200.5', ?, '2005', ?, '10',
			         '2026-03-30T00:31:00Z')`, id, newAvg, newAmount)
		return err
	}
	if err := insert("c-1", "200.6", "2006"); err != nil {
		t.Fatalf("first correction: %v", err)
	}
	if err := insert("c-2", "200.6", "2006"); err == nil {
		t.Fatal("re-observing the same correction must violate the UNIQUE constraint")
	}
	// A further correction at the same cumulative quantity is a distinct event.
	if err := insert("c-3", "200.7", "2007"); err != nil {
		t.Fatalf("a different correction at the same quantity must be recorded: %v", err)
	}
}

// TestExecutionCorrectionsOrderForeignKey keeps a correction attached to the
// snapshot it corrects: the prev values only mean something next to that row.
func TestExecutionCorrectionsOrderForeignKey(t *testing.T) {
	j := openTestJournal(t)
	_, err := j.db.ExecContext(context.Background(),
		`INSERT INTO execution_corrections
		   (id, account_ref, order_id, new_avg_price, new_filled_amount, cumulative_qty, observed_at)
		 VALUES ('c-1','acct-1','order-missing','1','1','1','2026-03-30T00:31:00Z')`)
	if err == nil {
		t.Fatal("a correction for an unknown order must violate the foreign key")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "foreign key") {
		t.Fatalf("want a foreign-key error, got %v", err)
	}
}

// TestExecutionCorrectionDedupColumnsAreNotNullable is why the dedup above
// actually holds: SQLite treats NULLs as distinct inside a UNIQUE constraint, so
// a nullable new_avg_price would let the same correction in twice whenever the
// broker reported no average price.
func TestExecutionCorrectionDedupColumnsAreNotNullable(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertFillSnapshot(t, j, "order-1")

	insert := func(id string) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO execution_corrections
			   (id, account_ref, order_id, new_avg_price, new_filled_amount,
			    cumulative_qty, observed_at)
			 VALUES (?, 'acct-1','order-1', NULL, NULL, '10','2026-03-30T00:31:00Z')`, id)
		return err
	}
	if err := insert("c-1"); err == nil {
		t.Fatal("new_avg_price/new_filled_amount must be NOT NULL so the dedup constraint bites")
	}

	// The unknown-value encoding is the empty string, as elsewhere in the
	// package (fill_snapshots.average_price), and it still dedups.
	insertBlank := func(id string) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO execution_corrections
			   (id, account_ref, order_id, cumulative_qty, observed_at)
			 VALUES (?, 'acct-1','order-1','10','2026-03-30T00:31:00Z')`, id)
		return err
	}
	if err := insertBlank("c-2"); err != nil {
		t.Fatalf("the empty-string default must be usable: %v", err)
	}
	if err := insertBlank("c-3"); err == nil {
		t.Fatal("two blank corrections at the same quantity must still dedup")
	}
}

// TestAttemptClientOrderIDUniquePerAccount pins D9's partial UNIQUE: an
// idempotency key is claimed by exactly one attempt per account, and attempts
// that carry no key (every pre-v5 row, and every path that cannot send one) are
// outside the index entirely.
func TestAttemptClientOrderIDUniquePerAccount(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertIntent(t, j, "intent-1")

	newAttempt := func(id, accountRef string, clientOrderID any) error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO mutation_attempts
			   (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at,
			    account_ref, client_order_id)
			 VALUES (?, 'intent-1', 'PLACE', 'RECORDED', 1, 'fp-1',
			         '2026-03-30T00:30:00Z', ?, ?)`, id, accountRef, clientOrderID)
		return err
	}

	if err := newAttempt("a-1", "acct-1", "key-1"); err != nil {
		t.Fatal(err)
	}
	if err := newAttempt("a-2", "acct-1", "key-1"); err == nil {
		t.Error("the same key twice on one account must be refused")
	}
	if err := newAttempt("a-3", "acct-2", "key-1"); err != nil {
		t.Errorf("the same key on another account is a different claim: %v", err)
	}
	// Keyless attempts are unconstrained: v4 rows have no key and CANCEL/AMEND
	// paths need not carry one.
	if err := newAttempt("a-4", "acct-1", nil); err != nil {
		t.Fatal(err)
	}
	if err := newAttempt("a-5", "acct-1", nil); err != nil {
		t.Errorf("two keyless attempts must both be allowed: %v", err)
	}
}

// TestPartialIndexPredicates reads the index definitions back out of
// sqlite_master. The WHERE clause is the difference between "unique among the
// rows that matter" and "unique always", and a behavioural test alone cannot
// tell a missing predicate from a test that never wrote the excluded rows.
func TestPartialIndexPredicates(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	want := map[string][]string{
		"idx_attempts_client_order": {
			"UNIQUE", "mutation_attempts", "account_ref", "client_order_id",
			"WHERE client_order_id IS NOT NULL",
		},
		"idx_reconcile_active": {
			"UNIQUE", "reconcile_states", "account_ref", "symbol",
			"WHERE released_at IS NULL",
		},
		"idx_reconcile_active_account_wide": {
			"UNIQUE", "reconcile_states", "account_ref",
			"WHERE released_at IS NULL AND symbol IS NULL",
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

// TestFillSnapshotFilledAmountDefault pins the column EXECUTION_CORRECTION needs
// as its "prev" (D7). It is NOT NULL and defaults to the empty string, like the
// average price beside it, so a v4 row migrated forward reads as "amount not
// observed yet" rather than as a NULL the comparison would have to special-case.
func TestFillSnapshotFilledAmountDefault(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertFillSnapshot(t, j, "order-1")

	var amount string
	if err := j.db.QueryRowContext(ctx,
		"SELECT filled_amount FROM fill_snapshots WHERE order_id = 'order-1'").Scan(&amount); err != nil {
		t.Fatal(err)
	}
	if amount != "" {
		t.Fatalf("filled_amount = %q, want the empty string (not observed)", amount)
	}
	if _, err := j.db.ExecContext(ctx,
		"UPDATE fill_snapshots SET filled_amount = NULL WHERE order_id = 'order-1'"); err == nil {
		t.Fatal("filled_amount is NOT NULL")
	}
}

// insertFillSnapshot writes a minimal snapshot row so the correction tests have
// an order to attach to.
func insertFillSnapshot(t *testing.T, j *Journal, orderID string) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO fill_snapshots (order_id, symbol, committed_at)
		 VALUES (?, 'AAPL', '2026-03-30T00:30:00Z')`, orderID); err != nil {
		t.Fatalf("insert fill snapshot %s: %v", orderID, err)
	}
}
