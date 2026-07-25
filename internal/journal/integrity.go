package journal

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrCorrupt means the journal database failed its integrity check. Startup is
// refused: the journal is the only record of what the engine was doing when it
// last ran, and acting on a damaged copy of that record can duplicate or lose a
// live order.
var ErrCorrupt = errors.New("journal: database integrity check failed")

// corruptionMarkers are the SQLite messages that mean "this file is not a healthy
// database" rather than "this operation failed".
var corruptionMarkers = []string{
	"database disk image is malformed",
	"file is not a database",
	"database corrupt",
	"malformed database schema",
	"encrypted or is not a database",
}

// checkIntegrity runs PRAGMA integrity_check and refuses anything but "ok".
//
// integrity_check (not quick_check) is used deliberately: startup happens once per
// engine run and the journal is small, so the full check is affordable and the
// weaker one can miss index damage that would later silently hide an attempt.
func (j *Journal) checkIntegrity(ctx context.Context) error {
	rows, err := j.db.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		if isCorruptionError(err) {
			return j.corruptError(err.Error())
		}
		return fmt.Errorf("journal: running the integrity check on %s: %w", j.path, err)
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			if isCorruptionError(err) {
				return j.corruptError(err.Error())
			}
			return fmt.Errorf("journal: reading the integrity check result for %s: %w", j.path, err)
		}
		messages = append(messages, line)
	}
	if err := rows.Err(); err != nil {
		if isCorruptionError(err) {
			return j.corruptError(err.Error())
		}
		return fmt.Errorf("journal: reading the integrity check result for %s: %w", j.path, err)
	}

	if len(messages) == 1 && strings.EqualFold(strings.TrimSpace(messages[0]), "ok") {
		return nil
	}
	return j.corruptError(strings.Join(messages, "; "))
}

// corruptError builds the refusal, including what an operator has to do next.
func (j *Journal) corruptError(detail string) error {
	if len(detail) > 512 {
		detail = detail[:512] + " …"
	}
	return fmt.Errorf(
		"%w: %s: %s\n"+
			"recovery: stop the engine, move %s (and its -wal/-shm files) aside, "+
			"reconcile open orders and positions against the broker account by hand, "+
			"then start again — or point %s at a restored copy. "+
			"Do not delete the file: it is the only record of the last run's mutations",
		ErrCorrupt, j.path, detail, j.path, EnvDataDir)
}

func isCorruptionError(err error) bool {
	msg := strings.ToLower(err.Error())
	for _, marker := range corruptionMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}
