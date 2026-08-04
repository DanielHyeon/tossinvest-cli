package strategyevidence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// OpenReadOnly opens an existing owner-only current-schema evidence store.
// It never creates a directory, database, schema, migration, WAL or snapshot.
func OpenReadOnly(ctx context.Context, options Options) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("strategy evidence: read-only context is required")
	}
	path := filepath.Clean(options.Path)
	owner, ownerOK := evidenceOwnerUID()
	if path == "." || path == "" || !filepath.IsAbs(path) || !ownerOK || validateEvidenceReadOnlyFile(path, owner) != nil {
		return nil, ErrSnapshotUnavailable
	}
	busy := options.BusyTimeout
	if busy <= 0 {
		busy = 250 * time.Millisecond
	}
	query := url.Values{}
	query.Set("mode", "ro")
	query.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busy.Milliseconds()))
	query.Add("_pragma", "query_only(1)")
	query.Add("_pragma", "foreign_keys(1)")
	uri := url.URL{Scheme: "file", Opaque: (&url.URL{Path: path}).EscapedPath()}
	db, err := sql.Open("sqlite", uri.String()+"?"+query.Encode())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	var version int
	if err := db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil || version != SchemaVersion {
		_ = db.Close()
		if version > SchemaVersion {
			return nil, ErrSchemaTooNew
		}
		return nil, ErrSnapshotUnavailable
	}
	clk := options.Clock
	if clk == nil {
		clk = marketclock.System()
	}
	return &Store{db: db, path: path, clk: clk}, nil
}
