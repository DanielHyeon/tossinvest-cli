package journal

// backup.go is the pre-migration copy and the way back to it.
//
// # Why a copy, and not a down-migration
//
// The migration rules in schema.go are additive: a step never drops or renames a
// column, so no step can undo itself. That is deliberate — a down-migration of a
// live account's order history is a data-loss tool — but it means an interrupted
// upgrade has no automatic way back. SQLite makes each step atomic (DDL inside a
// transaction rolls back like anything else), so the common failure leaves the
// journal exactly as it was. The copy exists for the failures SQLite cannot undo:
// a disk that fills between two steps, a kill in the middle of a multi-step run
// that already committed step N, an upgrade whose new code turns out to read the
// migrated file wrongly.
//
// # Restore procedure
//
// The failure message names the backup, and schema_meta.pre_migration_backup_path
// records it too. To go back:
//
//  1. Stop the engine and any tossctl process. Nothing may hold the journal open.
//  2. Move the three live files aside together — journal.db, journal.db-wal and
//     journal.db-shm. Do not delete them: a commit that was never checkpointed
//     lives only in the -wal, so the trio is the real state of the last run
//     (crash_linux_test.go's TestDetachedWALLosesCommittedRecord pins this).
//  3. Copy the backup into place as journal.db. It is self-contained — VACUUM
//     INTO checkpoints the WAL into the copy — so there is no -wal or -shm to
//     restore alongside it, and the ones you moved aside must not be put back.
//  4. chmod 600 the restored file: it holds a live account's order history.
//  5. Start the build that wrote the version in the backup. A newer build would
//     simply migrate it forward again, into whatever failed.
//  6. Before enabling entries, reconcile open orders and positions against the
//     broker by hand. The window between the backup and the failure may contain
//     mutations that the restored file does not know about.
//
// # Why VACUUM INTO rather than copying the file
//
// A plain file copy of journal.db drops every commit still sitting in the WAL,
// which under synchronous=FULL + WAL is most recent commits. VACUUM INTO writes a
// consistent, fully checkpointed, single-file copy from a read transaction, so
// the backup is a database rather than a fragment of one.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"
)

// backupTimeLayout stamps the backup name. RFC3339's colons are legal in a POSIX
// filename but awkward in every shell that has to move the file, so the compact
// form is used: 20260330T003000Z.
const backupTimeLayout = "20060102T150405Z"

// maxBackupSuffix bounds the collision search. Reaching it means the data
// directory already holds a hundred backups from the same instant, which is a
// broken loop rather than a situation to keep writing into.
const maxBackupSuffix = 99

// backupBeforeMigration copies the journal next to itself and returns the path.
//
// Failing to take the copy fails the migration: an upgrade that cannot be undone
// and cannot be restored is not one to start on a live account's journal.
func (j *Journal) backupBeforeMigration(ctx context.Context, from, to int) (string, error) {
	path, err := j.backupPath(from, to)
	if err != nil {
		return "", err
	}

	// VACUUM INTO cannot run inside a transaction; migrate calls this before the
	// first step opens one.
	if _, err := j.db.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return "", fmt.Errorf(
			"journal: backing up %s to %s before migrating v%d→v%d: %w\n"+
				"the migration was not started. Free space in the data directory, or point %s "+
				"at a filesystem with room, and try again",
			j.path, path, from, to, err, EnvDataDir)
	}
	// The copy holds everything the journal holds, so it inherits the same
	// permissions rather than the process umask's idea of them.
	if err := os.Chmod(path, 0o600); err != nil {
		return "", fmt.Errorf("journal: securing the pre-migration backup %s: %w", path, err)
	}

	// Recorded in the live database, not only in the error text: an operator who
	// finds the process dead has no error text to read.
	now := j.clk.Now().UTC().Format(time.RFC3339)
	for key, value := range map[string]string{
		"pre_migration_backup_path":    path,
		"pre_migration_backup_at":      now,
		"pre_migration_backup_version": strconv.Itoa(from),
	} {
		if _, err := j.db.ExecContext(ctx,
			`INSERT INTO schema_meta(key, value) VALUES(?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
			return "", fmt.Errorf("journal: recording the pre-migration backup path: %w", err)
		}
	}
	return path, nil
}

// backupPath builds a name that says what it is a backup of and when it was
// taken, and never picks a name that already exists: an earlier backup may be
// the only good state left, so overwriting one is the exact failure this file
// exists to prevent.
func (j *Journal) backupPath(from, to int) (string, error) {
	stamp := j.clk.Now().UTC().Format(backupTimeLayout)
	base := fmt.Sprintf("%s.v%d-pre-v%d.%s", j.path, from, to, stamp)

	for i := 0; i <= maxBackupSuffix; i++ {
		candidate := base + ".bak"
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d.bak", base, i)
		}
		_, err := os.Stat(candidate)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return candidate, nil
		case err != nil:
			return "", fmt.Errorf("journal: checking for an existing backup at %s: %w", candidate, err)
		}
	}
	return "", fmt.Errorf(
		"journal: %s already has %d pre-migration backups from the same instant — "+
			"move them aside before upgrading", base, maxBackupSuffix+1)
}

// withRestoreHint appends where the copy is and what to do with it. A migration
// failure is read by whoever is standing in front of a stopped engine, so the
// recovery has to be in the message rather than in a document they would have to
// go and find.
func (j *Journal) withRestoreHint(err error, backup string) error {
	if backup == "" {
		return err
	}
	return fmt.Errorf("%w\n"+
		"the journal was backed up before the migration started: %s\n"+
		"to restore: stop every engine/tossctl process, move %s and its -wal/-shm files aside "+
		"(do not delete them), copy the backup into place as %s, chmod 600 it, and run the build "+
		"that wrote that schema version. Reconcile open orders against the broker before enabling "+
		"entries — the full procedure is in internal/journal/backup.go",
		err, backup, j.path, j.path)
}
