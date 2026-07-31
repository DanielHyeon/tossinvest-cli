package main

import (
	"path/filepath"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// consoleJournalPath follows the engine's active-profile rule. An explicit
// config directory owns the journal beside its config; only the default profile
// uses the platform data directory.
func consoleJournalPath(root *rootOptions) (string, error) {
	if root != nil && root.configDir != "" {
		return filepath.Join(root.configDir, journal.DBFileName), nil
	}
	return journal.DefaultPath()
}
