package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	// modernc.org/sqlite is the pure-Go SQLite driver: no cgo, so the engine
	// cross-compiles and the release binary keeps working on a host without a
	// C toolchain. It registers itself as "sqlite".
	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// ErrSchemaTooNew means the journal on disk was written by a newer TossOS. An
// older binary must not touch it: it would read columns it does not know about as
// absent and could take an unsafe decision from an incomplete picture.
var ErrSchemaTooNew = errors.New("journal: database schema is newer than this build")

// defaultBusyTimeout bounds how long a writer waits for another writer's
// transaction. The journal has a single writer by design, so hitting this is a
// signal (a second engine instance, a stuck transaction), not routine contention.
const defaultBusyTimeout = 5 * time.Second

// Options configures Open.
type Options struct {
	// Path is the SQLite file. Empty resolves DefaultPath() from the
	// environment ($TOSSOS_DATA_DIR > $XDG_DATA_HOME/tossos > ~/.local/share).
	Path string

	// Clock stamps journal records. Defaults to clock.System(). Tests inject a
	// fake so recorded timestamps are deterministic.
	Clock clock.Clock

	// FSProber inspects the filesystem hosting the journal. Defaults to
	// SystemFSProber(); tests inject FixedFSProber because TMPDIR is not
	// necessarily on an allowlisted filesystem.
	FSProber FSProber

	// BusyTimeout overrides how long a blocked writer waits. Zero uses
	// defaultBusyTimeout.
	BusyTimeout time.Duration

	// migrationOverride replaces the forward-step list and the version this
	// build claims to understand. Unexported, therefore test-only: the migration
	// tests need a genuine older database on disk and a step that fails partway,
	// and neither can be expressed with the released migration list. Nil in
	// production, where defaultMigrationPlan applies.
	migrationOverride *migrationPlan
}

// Journal is an open journal database.
type Journal struct {
	db   *sql.DB
	path string
	clk  clock.Clock
	fs   FSInfo

	// applyHooks are the domain's injected apply functions, called inside the
	// fill transaction (apply_hook.go). Bound once at wiring time; the mutex is
	// there because binding happens on the wiring goroutine and the calls happen
	// on the detector's.
	applyMu    sync.RWMutex
	applyHooks ApplyHooks

	// modeProjector receives every committed operating-mode transition
	// (operating_mode.go). Bound once at wiring time, for the same reason the
	// apply hooks are: the projection is the enforcement point, and a second one
	// would be a second answer.
	modeMu        sync.RWMutex
	modeProjector ModeProjector

	// exitWriteHook is a same-package test seam for transaction fault injection.
	// Production never assigns it.
	exitWriteHook func(stage string) error
}

// Open resolves the path, verifies the filesystem, creates the data directory if
// needed, opens the database with the durability pragmas and migrates the schema
// forward.
//
// It refuses (returns an error, opens nothing) when the filesystem is outside the
// allowlist or the on-disk schema is newer than this build.
func Open(ctx context.Context, opts Options) (*Journal, error) {
	path := opts.Path
	if path == "" {
		resolved, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = resolved
	}
	path = filepath.Clean(path)
	dir := filepath.Dir(path)

	prober := opts.FSProber
	if prober == nil {
		prober = SystemFSProber()
	}
	// The guard runs before anything is created: a refused mount must be left
	// untouched.
	fs, err := CheckFilesystem(prober, dir)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, dataDirPerm); err != nil {
		return nil, fmt.Errorf("journal: creating %s: %w", dir, err)
	}

	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	busy := opts.BusyTimeout
	if busy <= 0 {
		busy = defaultBusyTimeout
	}

	db, err := sql.Open("sqlite", dsn(path, busy))
	if err != nil {
		return nil, fmt.Errorf("journal: opening %s: %w", path, err)
	}
	// One connection: the journal is a single-writer store, and serialising every
	// statement onto one connection is what makes "BEGIN IMMEDIATE then commit
	// then submit" hold without SQLITE_BUSY races between pooled connections.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("journal: connecting to %s: %w", path, err)
	}

	j := &Journal{db: db, path: path, clk: clk, fs: fs}
	// Corruption is checked before the schema is touched: a migration against a
	// damaged file would write on top of the damage. A corrupt journal means we
	// cannot say what the engine did last, so startup is refused rather than
	// continued from a record we know is wrong.
	if err := j.checkIntegrity(ctx); err != nil {
		db.Close()
		return nil, err
	}
	plan := defaultMigrationPlan()
	if opts.migrationOverride != nil {
		plan = *opts.migrationOverride
	}
	if err := j.migrate(ctx, plan); err != nil {
		db.Close()
		return nil, err
	}
	// Keep the journal private even if the process umask is permissive: it holds
	// the full order history of a live account.
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		if _, statErr := os.Stat(p); statErr == nil {
			_ = os.Chmod(p, 0o600)
		}
	}
	return j, nil
}

// dsn builds the driver DSN. The pragmas are the durability contract:
//
//	journal_mode=WAL   readers never block the writer; commits are crash-safe
//	synchronous=FULL   every commit fsyncs the WAL before returning
//	foreign_keys=ON    an attempt cannot exist without its intent
//	busy_timeout       bounded wait instead of an immediate SQLITE_BUSY
//	_txlock=immediate  database/sql's BeginTx issues BEGIN IMMEDIATE, taking the
//	                   write lock up front rather than discovering the conflict at
//	                   COMMIT time
func dsn(path string, busy time.Duration) string {
	q := url.Values{}
	q.Add("_pragma", "journal_mode(wal)")
	q.Add("_pragma", "synchronous(full)")
	q.Add("_pragma", "foreign_keys(on)")
	q.Add("_pragma", "busy_timeout("+strconv.FormatInt(busy.Milliseconds(), 10)+")")
	q.Set("_txlock", "immediate")
	return "file:" + path + "?" + q.Encode()
}

// Close releases the database handle.
func (j *Journal) Close() error {
	if j == nil || j.db == nil {
		return nil
	}
	return j.db.Close()
}

// Path returns the journal file path.
func (j *Journal) Path() string { return j.path }

// Filesystem returns what the guard found under the journal directory.
func (j *Journal) Filesystem() FSInfo { return j.fs }

// Now returns the injected clock's current instant.
func (j *Journal) Now() time.Time { return j.clk.Now() }

// SchemaVersion returns the schema version stored in the database.
func (j *Journal) SchemaVersion(ctx context.Context) (int, error) {
	var v int
	if err := j.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&v); err != nil {
		return 0, fmt.Errorf("journal: reading schema version: %w", err)
	}
	return v, nil
}

// migrate applies every pending forward step inside one transaction per step, and
// refuses to run against a newer schema.
//
// A journal that already holds rows is copied first (backup.go). The copy is
// taken before the first step and the migration is abandoned if it cannot be
// taken: the additive rules leave no way back down, so the copy is the only
// recovery from a step that dies partway, and starting without one would trade a
// live account's history for a faster startup.
func (j *Journal) migrate(ctx context.Context, plan migrationPlan) error {
	current, err := j.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if current > plan.target {
		return fmt.Errorf("%w: found version %d, this build understands %d — upgrade tossctl or point %s elsewhere",
			ErrSchemaTooNew, current, plan.target, EnvDataDir)
	}
	if current == plan.target {
		return nil
	}

	// current == 0 is a database this run just created: there is no history to
	// lose, and a backup of it would only litter the data directory.
	backup := ""
	if current > 0 {
		backup, err = j.backupBeforeMigration(ctx, current, plan.target)
		if err != nil {
			return err
		}
	}

	for _, m := range plan.steps {
		if m.Version <= current {
			continue
		}
		if err := j.applyMigration(ctx, m); err != nil {
			return j.withRestoreHint(err, backup)
		}
		current = m.Version
	}
	if current != plan.target {
		return j.withRestoreHint(
			fmt.Errorf("journal: migration ended at version %d, want %d", current, plan.target),
			backup)
	}
	return nil
}

func (j *Journal) applyMigration(ctx context.Context, m migration) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("journal: starting migration %d: %w", m.Version, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("journal: applying migration %d: %w", m.Version, err)
	}
	now := j.clk.Now().UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_meta(key, value) VALUES('schema_version', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		strconv.Itoa(m.Version)); err != nil {
		return fmt.Errorf("journal: recording schema version %d: %w", m.Version, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_meta(key, value) VALUES('created_at', ?)
		 ON CONFLICT(key) DO NOTHING`, now); err != nil {
		return fmt.Errorf("journal: recording created_at: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_meta(key, value) VALUES('migrated_at', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`, now); err != nil {
		return fmt.Errorf("journal: recording migrated_at: %w", err)
	}
	// PRAGMA user_version does not accept a bound parameter.
	if _, err := tx.ExecContext(ctx, "PRAGMA user_version = "+strconv.Itoa(m.Version)); err != nil {
		return fmt.Errorf("journal: setting schema version %d: %w", m.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: committing migration %d: %w", m.Version, err)
	}
	return nil
}
