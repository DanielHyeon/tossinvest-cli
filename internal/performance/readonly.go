package performance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrDatabaseMissing   = errors.New("performance: database is missing")
	ErrDatabaseWALActive = errors.New("performance: database has active WAL state")
	ErrDatabaseChanged   = errors.New("performance: database changed during immutable read")
	ErrSchemaTooOld      = errors.New("performance: database schema is older than this build")
)

type readOnlyFileIdentity struct {
	size    int64
	modTime time.Time
}

// ReadOnly exposes the derived dashboard query and lifecycle only. It has no
// Collect, Append, Prune, migration, journal, broker, or trading-state method.
// Each dashboard call takes a fresh immutable snapshot so a writer that becomes
// active after startup is detected instead of being silently ignored forever.
type ReadOnly struct {
	mu     sync.RWMutex
	path   string
	closed bool
}

// OpenReadOnly validates an existing clean checkpoint without creating its
// directory/file/sidecars, changing journal mode, running migrations, or
// acquiring a SQLite writer lock. Active WAL state is unavailable: immutable
// SQLite would ignore it, so accepting it could present stale evidence.
func OpenReadOnly(path string) (*ReadOnly, error) {
	path = filepath.Clean(path)
	db, identity, err := openImmutablePerformanceDB(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var version int
	if err := db.QueryRowContext(context.Background(), "PRAGMA user_version").Scan(&version); err != nil {
		return nil, fmt.Errorf("performance: reading schema version: %w", err)
	}
	switch {
	case version < SchemaVersion:
		return nil, fmt.Errorf("%w: found %d, require %d", ErrSchemaTooOld, version, SchemaVersion)
	case version > SchemaVersion:
		return nil, fmt.Errorf("%w: found %d, understand %d", ErrSchemaTooNew, version, SchemaVersion)
	}
	if err := unchangedImmutablePerformanceDB(path, identity); err != nil {
		return nil, err
	}
	return &ReadOnly{path: path}, nil
}

func openImmutablePerformanceDB(path string) (*sql.DB, readOnlyFileIdentity, error) {
	if strings.TrimSpace(path) == "" || path == "." {
		return nil, readOnlyFileIdentity{}, errors.New("performance: database path is required")
	}
	identity, err := cleanPerformanceDBIdentity(path)
	if err != nil {
		return nil, readOnlyFileIdentity{}, err
	}
	query := url.Values{}
	query.Set("mode", "ro")
	query.Set("immutable", "1")
	query.Add("_pragma", "query_only(true)")
	query.Add("_pragma", "busy_timeout("+strconv.FormatInt((5*time.Second).Milliseconds(), 10)+")")
	db, err := sql.Open("sqlite", "file:"+path+"?"+query.Encode())
	if err != nil {
		return nil, readOnlyFileIdentity{}, fmt.Errorf("performance: opening %s immutable read-only: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, readOnlyFileIdentity{}, fmt.Errorf("performance: connecting to %s immutable read-only: %w", path, err)
	}
	if err := unchangedImmutablePerformanceDB(path, identity); err != nil {
		db.Close()
		return nil, readOnlyFileIdentity{}, err
	}
	return db, identity, nil
}

func cleanPerformanceDBIdentity(path string) (readOnlyFileIdentity, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return readOnlyFileIdentity{}, fmt.Errorf("%w: %s", ErrDatabaseMissing, path)
		}
		return readOnlyFileIdentity{}, fmt.Errorf("performance: looking for %s: %w", path, err)
	}
	if info.IsDir() {
		return readOnlyFileIdentity{}, fmt.Errorf("%w: %s is a directory", ErrDatabaseMissing, path)
	}
	for _, sidecar := range []string{path + "-wal", path + "-shm"} {
		if _, err := os.Stat(sidecar); err == nil {
			return readOnlyFileIdentity{}, fmt.Errorf("%w: %s", ErrDatabaseWALActive, sidecar)
		} else if !errors.Is(err, os.ErrNotExist) {
			return readOnlyFileIdentity{}, fmt.Errorf("performance: checking WAL state %s: %w", sidecar, err)
		}
	}
	return readOnlyFileIdentity{size: info.Size(), modTime: info.ModTime()}, nil
}

func unchangedImmutablePerformanceDB(path string, before readOnlyFileIdentity) error {
	after, err := cleanPerformanceDBIdentity(path)
	if err != nil {
		return err
	}
	if before.size != after.size || !before.modTime.Equal(after.modTime) {
		return fmt.Errorf("%w: %s", ErrDatabaseChanged, path)
	}
	return nil
}

func (r *ReadOnly) Dashboard(ctx context.Context, query Query) (DashboardView, error) {
	if r == nil {
		return DashboardView{}, errors.New("performance: read-only database is not open")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return DashboardView{}, errors.New("performance: read-only database is closed")
	}
	db, identity, err := openImmutablePerformanceDB(r.path)
	if err != nil {
		return DashboardView{}, err
	}
	view, queryErr := (&Store{db: db, path: r.path}).Dashboard(ctx, query)
	closeErr := db.Close()
	if queryErr != nil {
		return DashboardView{}, queryErr
	}
	if closeErr != nil {
		return DashboardView{}, closeErr
	}
	if err := unchangedImmutablePerformanceDB(r.path, identity); err != nil {
		return DashboardView{}, err
	}
	return view, nil
}

// AttributionRows reads the exact persisted attribution generation through the
// same immutable, query-only snapshot discipline as Dashboard.
func (r *ReadOnly) AttributionRows(ctx context.Context, accountRef string, query AttributionQuery, limit int) ([]Attribution, error) {
	if r == nil {
		return nil, errors.New("performance: read-only database is not open")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return nil, errors.New("performance: read-only database is closed")
	}
	db, identity, err := openImmutablePerformanceDB(r.path)
	if err != nil {
		return nil, err
	}
	rows, queryErr := queryAttributionRows(ctx, db, accountRef, query, limit)
	closeErr := db.Close()
	if queryErr != nil {
		return nil, queryErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err := unchangedImmutablePerformanceDB(r.path, identity); err != nil {
		return nil, err
	}
	return rows, nil
}

// AttributionEvidence returns the canonical rebuild source through an
// immutable, query-only snapshot.
func (r *ReadOnly) AttributionEvidence(ctx context.Context, accountRef string) (AttributionRebuild, error) {
	if r == nil {
		return AttributionRebuild{}, errors.New("performance: read-only database is not open")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return AttributionRebuild{}, errors.New("performance: read-only database is closed")
	}
	db, identity, err := openImmutablePerformanceDB(r.path)
	if err != nil {
		return AttributionRebuild{}, err
	}
	evidence, queryErr := queryAttributionEvidence(ctx, db, accountRef)
	closeErr := db.Close()
	if queryErr != nil {
		return AttributionRebuild{}, queryErr
	}
	if closeErr != nil {
		return AttributionRebuild{}, closeErr
	}
	if err := unchangedImmutablePerformanceDB(r.path, identity); err != nil {
		return AttributionRebuild{}, err
	}
	return evidence, nil
}

func (r *ReadOnly) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	return nil
}

func (r *ReadOnly) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}
