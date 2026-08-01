package protection

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return openTestDBHandle(t, filepath.Join(t.TempDir(), "protection.db")+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", 8)
}

func openConcurrentTestDBs(t *testing.T) (*sql.DB, *sql.DB) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "protection-concurrent.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	return openTestDBHandle(t, dsn, 1), openTestDBHandle(t, dsn, 1)
}

func openTestDBHandle(t *testing.T, dsn string, maxConnections int) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(maxConnections)
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
