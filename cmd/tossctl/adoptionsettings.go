package main

// adoptionsettings.go wires the console's one config write surface (change
// console-adoption-controls task 3.4): a two-method seam over
// config.Service's raw load and surgical save, plus a save-time audit entry.
//
// The audit append lives HERE, not in internal/console: the console package
// writes nothing (its static guards say so), and §0.5 wants the flip recorded
// at the moment it is made — the engine's startup diff alone would collapse an
// on→off→on sequence between two starts into silence (review round 1, P2-9).
// The two records are deliberate: this one says when the person saved, the
// engine's says what the engine actually started with.

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// newAdoptionSettingsSeam resolves the config file the same way every other
// per-profile path is resolved: --config-dir wins, the default paths otherwise.
// The console.AdoptionSettings adaptation lives in console.go — the only file
// this command allows to import internal/console.
func newAdoptionSettingsSeam(root *rootOptions) *consoleAdoptionSettings {
	var path string
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		path = filepath.Join(root.configDir, "config.json")
	} else if paths, err := config.DefaultPaths(); err == nil {
		path = paths.ConfigFile
	} else {
		return nil
	}
	return &consoleAdoptionSettings{svc: config.NewService(path)}
}

type consoleAdoptionSettings struct{ svc *config.Service }

func (s consoleAdoptionSettings) Load() (config.Adoption, string, error) {
	return s.svc.LoadRawEngineAdoption()
}

func (s consoleAdoptionSettings) Save(a config.Adoption) error {
	if err := s.svc.SaveEngineAdoption(a); err != nil {
		return err
	}
	// Best-effort by design: the save is already durable and the engine's own
	// startup record is the second, mandatory trail. A console save must not be
	// rolled back because the audit disk write failed. RecordChange itself
	// de-duplicates against the last recorded value per setting.
	recordAdoptionSave(a)
	return nil
}

// recordAdoptionSave appends the save-time audit entries.
func recordAdoptionSave(next config.Adoption) {
	dir, err := journal.DataDir()
	if err != nil {
		return
	}
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, audit.FileName)})
	if err != nil {
		return
	}
	_, _ = log.RecordChange(audit.ActionAdoptionToggle, "console.adoption.enabled",
		strconv.FormatBool(next.Enabled), "saved from the console settings screen")
	_, _ = log.RecordChange(audit.ActionAdoptionSetting, "console.adoption.default_stop_pct",
		strconv.FormatFloat(next.DefaultStopPct, 'f', -1, 64), "")
	_, _ = log.RecordChange(audit.ActionAdoptionSetting, "console.adoption.exclude_symbols",
		strings.Join(next.ExcludeSymbols, ","), "")
	_, _ = log.RecordChange(audit.ActionAdoptionSetting, "console.adoption.include_symbols",
		strings.Join(next.IncludeSymbols, ","), "")
}
