// Package paths resolves where tossinvest keeps its config, session, lineage
// cache, official credentials and token cache, applying the CLI's
// --config-dir / --session-file overrides.
//
// It is a leaf package on purpose. Both wiring profiles need this resolution,
// but the engine profile (internal/app/engine) must not import the CLI profile
// (internal/app), which pulls in the hybrid router and the web-session order
// mutators. Keeping the resolution here lets both share one definition of "where
// the credentials live" while the engine's dependency graph stays free of any
// WTS mutator (proved by internal/app/engine's dependency-graph test).
package paths

import (
	"path/filepath"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// Resolve returns the default path set with the two CLI overrides applied.
//
// configDir moves the config file, session file and lineage cache into one
// directory — which is how tests keep a run away from the developer's real
// credentials. sessionFile overrides just the session path and wins over
// configDir.
func Resolve(configDir, sessionFile string) (config.Paths, error) {
	paths, err := config.DefaultPaths()
	if err != nil {
		return config.Paths{}, err
	}

	if configDir != "" {
		paths.ConfigDir = configDir
		paths.ConfigFile = filepath.Join(configDir, "config.json")
		paths.SessionFile = filepath.Join(configDir, "session.json")
		paths.LineageFile = filepath.Join(configDir, "trading-lineage.json")
	}

	if sessionFile != "" {
		paths.SessionFile = sessionFile
	}

	return paths, nil
}

// OpenAPI returns the official credentials file and token cache path, honouring
// a configDir override. When configDir is set both files are placed inside it so
// a caller can control them via a single temp directory.
func OpenAPI(configDir string) (credFile, tokenFile string, err error) {
	// Honour the override first so a DefaultPaths() failure never blocks callers
	// that already pin a config directory.
	if configDir != "" {
		return filepath.Join(configDir, "openapi-credentials.json"),
			filepath.Join(configDir, "openapi-token.json"),
			nil
	}
	paths, err := config.DefaultPaths()
	if err != nil {
		return "", "", err
	}
	return paths.CredentialsFile, paths.TokenFile, nil
}
