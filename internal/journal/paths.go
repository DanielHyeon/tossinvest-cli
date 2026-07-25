// Package journal is the durable record of every order intent and mutation
// attempt the engine makes.
//
// It exists because a broker with no idempotency key cannot tell us whether a
// mutation we failed to read the response for actually happened. The only way to
// answer that after a crash is to have written down what we were about to do
// *before* doing it, durably enough that the record survives power loss. That is
// this package's whole job, and it is why the rules below are strict:
//
//   - The journal lives outside the repository and outside the config directory,
//     under $TOSSOS_DATA_DIR > $XDG_DATA_HOME/tossos > ~/.local/share/tossos.
//   - It refuses to run on a filesystem where fsync does not mean fsync
//     (see fsguard.go — network, union, in-memory and FUSE-backed mounts).
//   - Intent writes commit with BEGIN IMMEDIATE + synchronous=FULL, and the
//     broker submission only happens after that commit returns.
//   - A corrupt journal refuses startup instead of degrading silently.
//
// Phase 2's ledger imports from here, so the schema keeps stable primary keys
// (intent id, attempt id), immutable intent fields and the full attempt history.
package journal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvDataDir is the operator override for the whole data directory.
	EnvDataDir = "TOSSOS_DATA_DIR"
	// EnvXDGDataHome is the XDG base directory variable.
	EnvXDGDataHome = "XDG_DATA_HOME"

	// AppDirName is the per-application subdirectory used under XDG data homes.
	AppDirName = "tossos"
	// DBFileName is the journal's SQLite file name inside the data directory.
	DBFileName = "journal.db"

	// dataDirPerm keeps the journal directory private to its owner: it contains
	// the complete order history of a live brokerage account.
	dataDirPerm = 0o700
)

var (
	// ErrRelativeDataDir is returned when $TOSSOS_DATA_DIR is set to a relative
	// path. An explicit operator override is not reinterpreted against the
	// current working directory — the journal would then move with the shell.
	ErrRelativeDataDir = errors.New("journal: " + EnvDataDir + " must be an absolute path")

	// ErrNoDataDir is returned when no override is set and the home directory
	// cannot be determined.
	ErrNoDataDir = errors.New("journal: cannot determine the data directory")
)

// DataDir resolves the directory that holds the journal, in the precedence the
// order-execution spec fixes:
//
//	$TOSSOS_DATA_DIR  →  $XDG_DATA_HOME/tossos  →  ~/.local/share/tossos
//
// A relative $XDG_DATA_HOME is ignored (as the XDG base directory spec requires);
// a relative $TOSSOS_DATA_DIR is an error.
func DataDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv(EnvDataDir)); override != "" {
		if !filepath.IsAbs(override) {
			return "", fmt.Errorf("%w (got %q)", ErrRelativeDataDir, override)
		}
		return filepath.Clean(override), nil
	}

	if xdg := strings.TrimSpace(os.Getenv(EnvXDGDataHome)); xdg != "" && filepath.IsAbs(xdg) {
		return filepath.Join(filepath.Clean(xdg), AppDirName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("%w: no %s, no usable %s and no home directory (%v)",
			ErrNoDataDir, EnvDataDir, EnvXDGDataHome, err)
	}
	return filepath.Join(home, ".local", "share", AppDirName), nil
}

// DefaultPath returns the journal database path inside the resolved data
// directory.
func DefaultPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DBFileName), nil
}
