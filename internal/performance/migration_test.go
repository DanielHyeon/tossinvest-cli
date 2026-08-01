package performance

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestFailedInitialMigrationRollsBackEveryDerivedTableAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	if _, err := openWithSchema(path, schemaV1+`INSERT INTO absent_table(x) VALUES(1);`, 1); err == nil {
		t.Fatal("broken schema unexpectedly opened")
	}
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version, tables int
	if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='performance_trades'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if version != 0 || tables != 0 {
		t.Fatalf("failed migration leaked version=%d tables=%d", version, tables)
	}
}

func TestPerformanceDatabaseFilesArePrivate(t *testing.T) {
	store := openTestStore(t)
	for _, path := range []string{store.Path(), store.Path() + "-wal", store.Path() + "-shm"} {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("%s mode=%#o want 0600", filepath.Base(path), got)
		}
	}
}

func TestStoreRefusesANewerDerivedSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA user_version = 2"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("newer derived schema unexpectedly opened")
	}
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRowContext(context.Background(), "PRAGMA user_version").Scan(&version); err != nil || version != 2 {
		t.Fatalf("newer version changed to %d err=%v", version, err)
	}
}
