package performance

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestOpenReadOnlyMissingDatabaseCreatesNothing(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "missing-data-dir")
	path := filepath.Join(dir, "performance.db")
	reader, err := OpenReadOnly(path)
	if !errors.Is(err, ErrDatabaseMissing) || reader != nil {
		t.Fatalf("OpenReadOnly missing=(%v,%v), want ErrDatabaseMissing and nil", reader, err)
	}
	if _, statErr := os.Stat(dir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("read-only open created data directory: %v", statErr)
	}
}

func TestOpenReadOnlyQueriesExistingDatabaseAndCannotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	writer, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Collect(context.Background(), measuredTrade(now.Add(-time.Hour)), nil, now); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, sidecar := range []string{path + "-wal", path + "-shm"} {
		if _, err := os.Stat(sidecar); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("writer fixture left sidecar %s: %v", sidecar, err)
		}
	}

	reader, err := OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	db, _, err := openImmutablePerformanceDB(path)
	if err != nil {
		t.Fatal(err)
	}
	view, err := reader.Dashboard(context.Background(), DefaultQuery(now))
	if err != nil || view.States.Complete != 1 || !view.NewestSourceAt.Equal(now.Add(-25*time.Minute)) {
		t.Fatalf("Dashboard=%+v err=%v", view, err)
	}
	if _, err := db.ExecContext(context.Background(), "CREATE TABLE forbidden_write(id INTEGER)"); err == nil {
		t.Fatal("query-only connection accepted a schema write")
	}
	if _, err := db.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err == nil {
		_, _ = db.ExecContext(context.Background(), "ROLLBACK")
		t.Fatal("read-only connection acquired a SQLite writer transaction")
	}
	var queryOnly int
	if err := db.QueryRowContext(context.Background(), "PRAGMA query_only").Scan(&queryOnly); err != nil || queryOnly != 1 {
		t.Fatalf("query_only=%d err=%v, want 1", queryOnly, err)
	}
	var found int
	if err := db.QueryRowContext(context.Background(),
		"SELECT count(*) FROM sqlite_master WHERE name='forbidden_write'").Scan(&found); err != nil || found != 0 {
		t.Fatalf("forbidden table count=%d err=%v", found, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	typ := reflect.TypeOf(reader)
	methods := make([]string, 0, typ.NumMethod())
	for i := 0; i < typ.NumMethod(); i++ {
		methods = append(methods, typ.Method(i).Name)
	}
	for _, forbidden := range []string{"Collect", "Prune", "AppendTrade", "AppendObservations"} {
		if slices.Contains(methods, forbidden) {
			t.Errorf("ReadOnly exposes writer method %s: %v", forbidden, methods)
		}
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("read-only open mutated database metadata: before=%d/%s after=%d/%s",
			before.Size(), before.ModTime(), after.Size(), after.ModTime())
	}
	for _, sidecar := range []string{path + "-wal", path + "-shm"} {
		if _, err := os.Stat(sidecar); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("immutable read created sidecar %s: %v", sidecar, err)
		}
	}
}

func TestOpenReadOnlyRefusesActiveWALRatherThanIgnoringIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	writer, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if _, err := os.Stat(path + "-wal"); err != nil {
		t.Fatalf("writer fixture has no active WAL: %v", err)
	}
	reader, err := OpenReadOnly(path)
	if !errors.Is(err, ErrDatabaseWALActive) || reader != nil {
		t.Fatalf("OpenReadOnly active WAL=(%v,%v), want fail-closed", reader, err)
	}
}

func TestOpenReadOnlyRejectsOldSchemaWithoutMigrating(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA user_version=0"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenReadOnly(path)
	if !errors.Is(err, ErrSchemaTooOld) || reader != nil {
		t.Fatalf("OpenReadOnly old=(%v,%v), want ErrSchemaTooOld and nil", reader, err)
	}
	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version, tables int
	if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := raw.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table'").Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if version != 0 || tables != 0 {
		t.Fatalf("read-only open migrated schema: version=%d tables=%d", version, tables)
	}
}

func TestDashboardNewestSourceAtDoesNotLaunderFreshnessWithCalculationTime(t *testing.T) {
	store := openTestStore(t)
	entry := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(entry)
	firstCalculated := entry.Add(time.Hour)
	secondCalculated := entry.Add(2 * time.Hour)
	if _, err := store.Collect(context.Background(), trade, nil, firstCalculated); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Collect(context.Background(), trade, nil, secondCalculated); err != nil {
		t.Fatal(err)
	}
	query := DefaultQuery(secondCalculated.Add(30 * time.Minute))
	query.MinimumSample = 1
	view, err := store.Dashboard(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	wantSource := trade.ClosedAt.UTC()
	if !view.Query.AsOf.Equal(query.AsOf) || !view.NewestSourceAt.Equal(wantSource) {
		t.Fatalf("query/source=%+v/%s want as-of %s authoritative source %s", view.Query, view.NewestSourceAt, query.AsOf, wantSource)
	}
}
