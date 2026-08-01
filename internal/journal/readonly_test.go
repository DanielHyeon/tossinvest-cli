package journal

// readonly_test.go pins the read-only access path (change add-operator-dashboard,
// task 1.1).
//
// The claims are not about convenience. The operator console opens the engine's
// journal while the engine is running, and every one of these tests is a way that
// access could quietly become a write:
//
//	nothing is created   a console started before the engine must not leave a
//	                     zero-row database behind that the engine then migrates
//	                     into the real one.
//	nothing is migrated  a console from an older build must not touch the schema.
//	nothing is written   mode=ro is the connection-level refusal; the absence of
//	                     write methods on the handle is the compile-time one.
//	both directions      a schema newer than this build and a schema older than
//	                     the tables this build reads are different problems with
//	                     different answers, and neither is an empty screen.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func openTestReadOnly(t *testing.T, path string) *ReadOnly {
	t.Helper()
	ro, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly(%s): %v", path, err)
	}
	t.Cleanup(func() {
		if err := ro.Close(); err != nil {
			t.Errorf("ReadOnly.Close: %v", err)
		}
	})
	return ro
}

// TestOpenReadOnlyCreatesNeitherFileNorDirectory.
//
// This is the whole reason OpenReadOnly exists rather than Open with a flag:
// Open does MkdirAll and then migrates, so a console that opened the journal
// before the engine ever ran would create the data directory and a v6 database
// with no rows in it. The engine would then start against a file somebody else
// made.
func TestOpenReadOnlyCreatesNeitherFileNorDirectory(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "not-yet")
	path := filepath.Join(dir, DBFileName)

	_, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrJournalMissing) {
		t.Fatalf("OpenReadOnly on a missing database = %v, want ErrJournalMissing", err)
	}
	if _, statErr := os.Stat(dir); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("the data directory was created: %v", statErr)
	}

	// And with the directory present but the database absent.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err = OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrJournalMissing) {
		t.Fatalf("OpenReadOnly on a missing database = %v, want ErrJournalMissing", err)
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("reading %s: %v", dir, readErr)
	}
	if len(entries) != 0 {
		t.Errorf("OpenReadOnly left %d file(s) behind: %v", len(entries), entries)
	}
}

// TestOpenReadOnlyRefusesToWriteAnything.
//
// mode=ro is the connection saying no. It is asserted here rather than trusted,
// because the DSN is a string and a string is exactly the kind of thing an edit
// can lose a word out of.
func TestOpenReadOnlyRefusesToWriteAnything(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	writable := openTestJournalAt(t, path)
	insertPosition(t, writable, "p-1", nil)

	ro := openTestReadOnly(t, path)
	ctx := context.Background()

	for _, stmt := range []string{
		`INSERT INTO schema_meta(key, value) VALUES('x','y')`,
		`UPDATE positions SET quantity = '0'`,
		`DELETE FROM positions`,
		`CREATE TABLE scratch (a TEXT)`,
		`PRAGMA user_version = 99`,
	} {
		if _, err := ro.db.ExecContext(ctx, stmt); err == nil {
			t.Errorf("a read-only connection executed %q", stmt)
		}
	}
}

// TestTheReadOnlyHandleHasNoWriteMethods.
//
// The runtime refusal above is the second lock. This is the first: a caller
// holding a *ReadOnly cannot reach a statement that writes, because no method
// exists that would take one. The allowlist is spelled out so that adding a
// method is a decision somebody has to make here, in the open.
func TestTheReadOnlyHandleHasNoWriteMethods(t *testing.T) {
	allowed := map[string]bool{
		"Close":             true,
		"Path":              true,
		"SchemaVersion":     true,
		"AccountRefs":       true,
		"LivePositionExits": true,
		"AccountExitEvents": true,
		"AccountTradeTrips": true,
		// The orders screen's origin join (change console-orders-screen, task
		// 2.1). It is a SELECT DISTINCT over one column of mutation_attempts and
		// it reaches the same mode=ro connection every other read here does.
		"BrokerOrderIDs": true,
		// a043's deterministic broker-order -> attempt -> exit-intent read. It is
		// another SELECT-only projection on the same query-only connection.
		"BrokerOrderExitLinks": true,
		// a049's exact, account-scoped SELECT over frozen outcomes and a047
		// lineage. Missing lineage remains nullable; this method has no writer.
		"ClosedStrategyTradeSources": true,
	}
	typ := reflect.TypeOf(&ReadOnly{})
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		if !allowed[name] {
			t.Errorf("*ReadOnly has an unlisted method %q; the console's journal handle is read-only "+
				"and every addition to it has to be argued for here", name)
		}
	}
}

// TestOpenReadOnlyDistinguishesTheTwoSchemaDirections.
//
// A journal newer than this build and a journal older than the tables this build
// reads are opposite problems: one wants a newer console, the other wants the
// engine to start and migrate. Reporting either as "no data" would be the console
// lying about an account.
func TestOpenReadOnlyDistinguishesTheTwoSchemaDirections(t *testing.T) {
	ctx := context.Background()

	t.Run("newer than this build", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), DBFileName)
		j := openTestJournalAt(t, path)
		if _, err := j.db.ExecContext(ctx, "PRAGMA user_version = 9999"); err != nil {
			t.Fatalf("bumping the version: %v", err)
		}

		_, err := OpenReadOnly(ctx, ReadOnlyOptions{Path: path})
		if !errors.Is(err, ErrSchemaTooNew) {
			t.Fatalf("OpenReadOnly against a newer schema = %v, want ErrSchemaTooNew", err)
		}
	})

	t.Run("older than the tables this build reads", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), DBFileName)
		// A v1 journal: real, valid, and without a single table the dashboard reads.
		j, err := Open(ctx, Options{
			Path:     path,
			Clock:    clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)),
			FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
			migrationOverride: &migrationPlan{
				steps:  []migration{{Version: 1, SQL: schemaV1}},
				target: 1,
			},
		})
		if err != nil {
			t.Fatalf("Open at v1: %v", err)
		}
		if err := j.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		_, err = OpenReadOnly(ctx, ReadOnlyOptions{Path: path})
		if !errors.Is(err, ErrSchemaTooOld) {
			t.Fatalf("OpenReadOnly against a pre-v6 schema = %v, want ErrSchemaTooOld", err)
		}
		// The message names what is missing, because "run the engine" is only
		// actionable if the operator can see why.
		if !strings.Contains(err.Error(), "positions") {
			t.Errorf("the refusal does not name a missing table: %v", err)
		}
	})
}

func TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	v14 := openJournalAtSchema(t, path, 14)
	if err := v14.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrSchemaTooOld) || !strings.Contains(err.Error(), "trade_outcomes.cost_total") {
		t.Fatalf("OpenReadOnly against v14=%v, want typed missing-cost prerequisite", err)
	}
}

// TestOpenReadOnlyDoesNotMigrate.
//
// The version on disk is the version after the read. An older console that
// migrated the journal forward would have rewritten a live account's history to
// suit itself.
func TestOpenReadOnlyDoesNotMigrate(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	writable := openTestJournalAt(t, path)
	ctx := context.Background()

	before, err := writable.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	ro := openTestReadOnly(t, path)
	if got := ro.SchemaVersion(); got != before {
		t.Errorf("the read-only open changed the schema version: %d → %d", before, got)
	}
	after, err := writable.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if after != before {
		t.Errorf("the schema version on disk moved: %d → %d", before, after)
	}
}

// TestOpenReadOnlyReadsWhatTheEngineIsWriting.
//
// The console opens the journal of a running engine. WAL is what makes that
// possible, and the read-only connection's access to the -shm/-wal pair is the
// documented exception to "creates nothing" — it is the precondition of reading a
// WAL database at all.
func TestOpenReadOnlyReadsWhatTheEngineIsWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	writable := openTestJournalAt(t, path)
	insertDecision(t, writable, "d-1", "nonce-1")
	insertPosition(t, writable, "p-1", "d-1")

	ro := openTestReadOnly(t, path)
	rows, err := ro.LivePositionExits(context.Background(), "acct-1")
	if err != nil {
		t.Fatalf("LivePositionExits: %v", err)
	}
	if len(rows) != 1 || rows[0].Position.ID != "p-1" {
		t.Fatalf("the read-only handle did not see the writer's row: %+v", rows)
	}

	// A row written after the read-only handle was opened is visible to the next
	// query on it: the console re-reads per request and must not be pinned to a
	// snapshot from startup.
	insertPosition2(t, writable, "p-2", "d-1", 2)
	rows, err = ro.LivePositionExits(context.Background(), "acct-1")
	if err != nil {
		t.Fatalf("LivePositionExits: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("a later write is invisible to the read-only handle: %+v", rows)
	}
}

// --- the order-origin read (change console-orders-screen, tasks 2.1-2.3) --------

// insertAttemptWithBrokerOrder records one PLACE attempt the broker acked.
func insertAttemptWithBrokerOrder(t *testing.T, j *Journal, attemptID, intentID, brokerOrderID string) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(),
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, broker_order_id, fingerprint, recorded_at)
		 VALUES (?, ?, 'PLACE', 'CONFIRMED', 1, ?, 'fp', '2026-03-30T00:30:00Z')`,
		attemptID, intentID, brokerOrderID); err != nil {
		t.Fatalf("insert attempt %s: %v", attemptID, err)
	}
}

// TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery.
//
// readOnlyTables exists for exactly this. Without the table registered,
// OpenReadOnly succeeds and the query above fails one statement at a time — and a
// failed query that the screen turns into zero rows reads as "every order was
// placed by hand". The refusal has to happen once, at open, where it can be named.
func TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	j := openTestJournalAt(t, path)
	// The engine's own foreign keys point at this table, so a real journal cannot
	// lose it — which is why the guard is worth having: the failure it catches is
	// a hand-built or partially-restored file, and that is the one nobody would
	// think to check before believing the screen.
	for _, stmt := range []string{
		"PRAGMA foreign_keys = OFF",
		"DROP TABLE attempt_transitions",
		"DROP TABLE mutation_attempts",
	} {
		if _, err := j.db.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	if err := j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrSchemaTooOld) {
		t.Fatalf("OpenReadOnly on a journal with no mutation_attempts = %v, want ErrSchemaTooOld", err)
	}
	if !strings.Contains(err.Error(), "mutation_attempts") {
		t.Errorf("the refusal does not name the missing table: %v", err)
	}
}
