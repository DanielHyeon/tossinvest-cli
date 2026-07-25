// Package app builds the wiring every tossinvest surface needs: resolved
// paths, loaded config, the session, the hybrid (WTS + official) client, the
// lineage recorder and the trading-policy service.
//
// It exists because that wiring used to live inside cmd/tossctl as
// newAppContext, where nothing but cobra could reach it. TossOS runs a trading
// engine in the same repository, and an engine that reconstructed the wiring
// itself would be free to disagree with the CLI about which paths hold the
// session or which backend may place orders — exactly the class of divergence
// that produces surprise live orders.
//
// Two profiles are provided and they are deliberately different:
//
//   - New (this file) is the CLI/MCP profile. It reproduces upstream behaviour
//     bit for bit, including the hybrid router that degrades to the web session
//     when official credentials are absent.
//   - the engine profile (internal/app/engine) constructs the official Open API
//     client directly and refuses to start without credentials. It never routes
//     order mutations through the web session.
package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JungHoonGhae/tossinvest-cli/internal/auth"
	tossclient "github.com/JungHoonGhae/tossinvest-cli/internal/client"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/hybrid"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderlineage"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Options are the run-scoped inputs the CLI takes from its persistent flags.
// The zero value is meaningful for every field except OutputFormat, which is
// parsed strictly (as the --output flag always is).
type Options struct {
	// OutputFormat is the --output value: "table", "json" or "csv".
	OutputFormat string
	// ConfigDir overrides the config directory (--config-dir). When set, the
	// config file, session file, lineage cache, credentials file and token
	// cache all resolve inside it, which is how tests keep a run away from the
	// developer's real credentials.
	ConfigDir string
	// SessionFile overrides just the session path (--session-file). It wins over
	// ConfigDir.
	SessionFile string
	// Backend overrides the configured routing preference for this run
	// (--backend): "auto", "wts" or "openapi". Empty means "use config".
	Backend string
}

// Context is the assembled wiring. Fields are exported because the surfaces
// that consume it live in other packages; it is a bag of dependencies, not an
// abstraction.
type Context struct {
	Format         output.Format
	Paths          config.Paths
	Config         config.File
	ConfigService  *config.Service
	LoginConfig    auth.LoginConfig
	AuthService    *auth.Service
	Client         *hybrid.Client
	Session        *session.Session
	TokenFile      string
	LineageService *orderlineage.Service
	TradingService *trading.Service
}

// New assembles the CLI/MCP profile.
//
// Behaviour is upstream's newAppContext unchanged, including the parts that
// look permissive:
//   - a missing session is not an error (official-only commands need none),
//   - a missing credentials file is not an error (the hybrid router serves
//     everything from the web session instead).
//
// A malformed config file or an unparseable flag value IS an error: silently
// falling back to defaults would flip the trading toggles to their zero values
// without telling the user.
func New(opts Options) (*Context, error) {
	format, err := output.ParseFormat(opts.OutputFormat)
	if err != nil {
		return nil, err
	}

	paths, err := ResolvePaths(opts.ConfigDir, opts.SessionFile)
	if err != nil {
		return nil, err
	}

	store := session.NewFileStore(paths.SessionFile)
	sess, err := store.Load(context.Background())
	if err != nil && !errors.Is(err, session.ErrNoSession) {
		return nil, err
	}

	loginConfig := auth.DefaultLoginConfig(paths.CacheDir)
	configService := config.NewService(paths.ConfigFile)
	cfg, err := configService.Load(context.Background())
	if err != nil {
		return nil, err
	}

	wtsClient := tossclient.New(tossclient.Config{
		Session:       sess,
		TradingPolicy: cfg.Trading,
	})

	prefer, err := ResolveBackend(cfg.OpenAPI, opts.Backend)
	if err != nil {
		return nil, err
	}

	credFile, tokenFile, err := ResolveOpenAPIPaths(opts.ConfigDir)
	if err != nil {
		return nil, err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return nil, fmt.Errorf("loading official credentials: %w", err)
	}

	var off *official.Client
	if creds != nil && cfg.OpenAPI.Enabled && prefer != "wts" {
		off = official.New(*creds, tokenFile)
	}

	h := hybrid.New(wtsClient, off, hybrid.Policy{Prefer: prefer, Fallback: cfg.OpenAPI.Fallback}, os.Stderr)

	lineage := orderlineage.NewService(paths.LineageFile)
	return &Context{
		Format:        format,
		Paths:         paths,
		Config:        cfg,
		ConfigService: configService,
		LoginConfig:   loginConfig,
		AuthService: auth.NewService(store, paths.SessionFile, auth.Options{
			LoginConfig:     loginConfig,
			Validator:       wtsClient,
			ExtensionRunner: wtsClient,
		}),
		Client:         h,
		Session:        sess,
		TokenFile:      tokenFile,
		LineageService: lineage,
		// The trading service records lineage itself, so every surface that
		// mutates through it (cobra, MCP, `ops call`) leaves the same trail.
		TradingService: trading.NewService(cfg.Trading, h.Broker()).WithLineage(lineage),
	}, nil
}

// ResolvePaths applies the --config-dir and --session-file overrides to the
// default path set. Exported because the CLI resolves individual paths before it
// has a Context (cobra's PersistentPreRunE runs before subcommands build one).
func ResolvePaths(configDir, sessionFile string) (config.Paths, error) {
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

// ResolveOpenAPIPaths returns the official credentials file and token cache
// path, honouring a --config-dir override. When configDir is set, both files
// are placed inside it so a test can control them via a single temp directory.
func ResolveOpenAPIPaths(configDir string) (credFile, tokenFile string, err error) {
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

// ResolveBackend returns the effective routing backend preference.
// The --backend flag takes precedence over cfg.Prefer.
// An empty flag means "use config". Invalid flag values are rejected.
func ResolveBackend(cfg config.OpenAPI, flag string) (string, error) {
	if flag == "" {
		return cfg.Prefer, nil
	}
	if norm, ok := config.NormalizeBackend(flag); ok {
		return norm, nil
	}
	return "", fmt.Errorf("invalid --backend value %q: must be one of auto, wts, openapi", flag)
}
