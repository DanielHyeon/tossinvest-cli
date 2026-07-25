// Package engine builds the order-execution wiring for TossOS's automated
// trading engine.
//
// It is a sibling of the CLI profile (internal/app), not a variant of it, and
// the difference is the point:
//
//   - The CLI is interactive and permissive. Without official credentials its
//     hybrid router serves orders from the scraped web session instead, and the
//     user's openapi.enabled / openapi.prefer settings decide the routing.
//   - The engine has no user watching it. TossOS declares the web session
//     read-only for order execution, so the engine constructs the official Open
//     API client directly, ignores openapi.enabled / openapi.prefer entirely,
//     and refuses to start at all when credentials are missing. Degrading
//     silently to WTS is the failure mode this package exists to make
//     impossible.
//
// The prohibition is structural, not just conventional: this package does not
// import internal/hybrid or internal/client, and a dependency-graph test
// (deps_test.go) fails the build if that ever changes — a web-session order
// mutator is not merely unused here, it is unspellable.
package engine

import (
	"context"
	"errors"
	"fmt"
	"os"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// ErrOfficialCredentialsRequired is returned by New when no usable official
// Open API credentials are available. It is a startup refusal, not a warning:
// there is no fallback broker for the engine.
var ErrOfficialCredentialsRequired = errors.New(
	"official Open API credentials are required to start the engine: run `tossctl openapi login` " +
		"or set TOSSCTL_OPENAPI_KEY and TOSSCTL_OPENAPI_SECRET (the engine never falls back to the web session)")

// Options are the engine's startup inputs.
type Options struct {
	// ConfigDir overrides the config directory, exactly as the CLI's
	// --config-dir does. Empty uses the per-user default.
	ConfigDir string
	// Getenv is the environment lookup used for credentials. Defaults to
	// os.Getenv; tests inject their own.
	Getenv func(string) string
	// OfficialOptions are passed to official.New. Tests use them to point the
	// client at an httptest server; production leaves this empty.
	OfficialOptions []official.Option
}

// Context is the engine's assembled wiring.
type Context struct {
	Paths     config.Paths
	Config    config.File
	TokenFile string

	// Official is the one broker connection the engine has.
	Official *official.Client
	// Broker is the official-only trading.Broker the trading service mutates
	// through.
	Broker trading.Broker
	// Conditional is the official-only conditional-order surface.
	Conditional ConditionalMutator
	// TradingService applies the user's config policy and the confirm-token gate
	// on top of Broker.
	TradingService *trading.Service
}

// ConditionalMutator is the conditional-order write surface. Its shape matches
// the methods *hybrid.Client exposes to the CLI, so both profiles speak the same
// intent types.
type ConditionalMutator interface {
	CreateConditionalOrder(ctx context.Context, intent orderintent.ConditionalPlaceIntent) (domain.ConditionalOrderRef, error)
	CancelConditionalOrder(ctx context.Context, intent orderintent.ConditionalCancelIntent) error
	ModifyConditionalOrder(ctx context.Context, intent orderintent.ConditionalModifyIntent) error
}

// New assembles the engine profile, or refuses to start.
//
// Refusals (all fail-closed, no partial engine is returned):
//   - no credentials, or credentials missing either half → ErrOfficialCredentialsRequired
//   - unreadable/malformed credentials file → the underlying error
//   - malformed config file → the underlying error, rather than running on
//     zero-value trading policy the user never wrote
//
// cfg.OpenAPI is loaded but never consulted: it is an interactive routing
// preference, and the engine's broker is not negotiable.
func New(opts Options) (*Context, error) {
	getenv := opts.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}

	paths, err := apppaths.Resolve(opts.ConfigDir, "")
	if err != nil {
		return nil, err
	}

	cfg, err := config.NewService(paths.ConfigFile).Load(context.Background())
	if err != nil {
		return nil, err
	}

	credFile, tokenFile, err := apppaths.OpenAPI(opts.ConfigDir)
	if err != nil {
		return nil, err
	}
	creds, err := official.LoadCredentials(getenv, credFile)
	if err != nil {
		return nil, fmt.Errorf("loading official credentials: %w", err)
	}
	// A file that exists but carries an empty key or secret is "configured" as
	// far as the CLI is concerned; for the engine it is unusable, and finding
	// that out on the first live order is too late.
	if creds == nil || creds.APIKey == "" || creds.SecretKey == "" {
		return nil, ErrOfficialCredentialsRequired
	}

	off := official.New(*creds, tokenFile, opts.OfficialOptions...)
	broker := &officialBroker{off: off}

	return &Context{
		Paths:       paths,
		Config:      cfg,
		TokenFile:   tokenFile,
		Official:    off,
		Broker:      broker,
		Conditional: broker,
		// No lineage recorder: the engine records order lineage in the journal
		// transaction (design D2), not in the CLI's JSON cache. Until the journal
		// lands, trading.Service without a recorder behaves exactly as upstream's
		// does when lineage is unset.
		TradingService: trading.NewService(cfg.Trading, broker),
	}, nil
}
