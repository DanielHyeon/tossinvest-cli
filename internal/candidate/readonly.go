package candidate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// OpenReadOnly opens an existing current-schema discovery store with SQLite's
// query_only and mode=ro controls. It never creates a directory, database,
// schema, migration, WAL or approval record.
func OpenReadOnly(ctx context.Context, opts Options) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("candidate: read-only context is required")
	}
	path := opts.Path
	if path == "" {
		resolved, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = resolved
	}
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("candidate: read-only store must be an existing regular file: %w", err)
	}
	prober := opts.FSProber
	if prober == nil {
		prober = SystemFSProber()
	}
	fs, err := CheckFilesystem(prober, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	busy := opts.BusyTimeout
	if busy <= 0 {
		busy = defaultBusyTimeout
	}
	db, err := sql.Open("sqlite", readOnlyDSN(path, busy))
	if err != nil {
		return nil, fmt.Errorf("candidate: opening read-only %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	var found string
	if err := db.QueryRowContext(ctx, `SELECT value FROM store_meta WHERE key='schema_version'`).Scan(&found); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("candidate: reading read-only schema version: %w", err)
	}
	var version int
	if _, err := fmt.Sscanf(found, "%d", &version); err != nil || version != SchemaVersion {
		_ = db.Close()
		if version > SchemaVersion {
			return nil, fmt.Errorf("%w (found %d, this build understands %d)", ErrSchemaTooNew, version, SchemaVersion)
		}
		return nil, fmt.Errorf("candidate: read-only store schema %q is not current version %d", found, SchemaVersion)
	}
	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	cooling := opts.CoolingTTL
	if cooling <= 0 {
		cooling = DefaultCoolingTTL
	}
	staleness := opts.StalenessTTL
	if staleness <= 0 {
		staleness = DefaultStalenessTTL
	}
	return &Store{db: db, path: path, clk: clk, fs: fs, prober: prober, cooling: cooling, staleness: staleness}, nil
}

func readOnlyDSN(path string, busy time.Duration) string {
	query := url.Values{}
	query.Set("mode", "ro")
	query.Add("_pragma", "busy_timeout("+fmt.Sprint(busy.Milliseconds())+")")
	query.Add("_pragma", "query_only(1)")
	query.Add("_pragma", "foreign_keys(1)")
	uri := url.URL{Scheme: "file", Opaque: (&url.URL{Path: path}).EscapedPath()}
	return uri.String() + "?" + query.Encode()
}
