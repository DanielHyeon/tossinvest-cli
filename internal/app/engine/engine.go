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
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
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

	// --- automation gate interlock (task 4.2) -------------------------------
	//
	// All of these are inert while the gate is off, which is the default and is
	// what every existing config produces.

	// Guardian is the risk authority that authorises mutations. Required when
	// the automation gate is on; ignored when it is off.
	Guardian execgw.Guardian
	// Clock drives the attestation's expiry check. Defaults to clock.System();
	// tests inject a fake so "expired" is a decision and not a race.
	Clock clock.Clock
	// AuditFile overrides where operational-setting changes are recorded. Empty
	// resolves <data dir>/audit.log — outside the config directory, because the
	// audit trail must survive a config directory being replaced.
	AuditFile string
	// Operator names who is starting the engine, for the audit trail. Empty
	// resolves the OS user.
	Operator string
}

// Context is the engine's assembled wiring.
type Context struct {
	Paths     config.Paths
	Config    config.File
	TokenFile string

	// Official is the engine's read-only view of its one broker connection.
	//
	// It is an interface, not *official.Client, and that is the seal: the
	// concrete client can place, cancel and modify orders, so a caller holding
	// the wiring could bypass the journal and the Guardian entirely. OfficialReads
	// (reads.go) declares no mutating method, so that call is not one the engine's
	// API can express.
	Official OfficialReads

	// Automation reports what the startup interlock decided about the automation
	// gate. Zero value = gate off.
	Automation AutomationStatus
	// Guardian is the injected risk authority, non-nil only when the gate is on
	// and verified.
	Guardian execgw.Guardian
	// Audit is the operational-settings audit log.
	Audit *audit.Log

	// official is the concrete client behind Official. It stays unexported: the
	// engine's own wiring occasionally needs the concrete type, and handing it
	// out would undo the seal the Official field exists to be.
	official *official.Client

	// broker and conditional are the official-only mutators. They are unexported
	// because internal/execgw's ExecutionGateway is now the engine's only order
	// path (engine-safety: "Guardian 결정 없는 제출 경로는 컴파일·API 수준에서
	// 존재하지 않아야 한다"). Handing either of them back as a field would be
	// exactly such a path: a caller holding a trading.Broker can place an order
	// with no GuardianDecision, no journal record and no IN_DOUBT handling.
	//
	// They stay reachable where they must be — TradingService wraps both, and the
	// gateway wraps TradingService.
	broker      trading.Broker
	conditional ConditionalMutator

	// TradingService applies the user's config policy and the confirm-token gate
	// on top of the broker. It is what internal/execgw wraps.
	TradingService *trading.Service
}

// ConditionalMutator is the conditional-order write surface. It is an alias of
// trading.ConditionalBroker (the gate and intent assembly live there, task 1.4),
// so the engine's adapter and the CLI's hybrid client are interchangeable at
// this seam.
type ConditionalMutator = trading.ConditionalBroker

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
//
// With config's automation gate on, four more refusals apply — no Guardian, zero
// limits, no/expired/mismatched capability attestation — all in interlock.go. With
// the gate off (the default, and every config written before schema v5) New makes
// no network call and behaves exactly as it did before the interlock existed.
func New(opts Options) (*Context, error) {
	return NewContext(context.Background(), opts)
}

// NewContext is New with a caller-supplied context. The context is used only by
// the automation-gate interlock's account read; with the gate off nothing in this
// function performs I/O over the network.
func NewContext(ctx context.Context, opts Options) (*Context, error) {
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

	auditLog, err := openAuditLog(opts)
	if err != nil {
		return nil, err
	}
	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}

	// The interlock runs before the context is returned, so a refused gate
	// produces no engine at all rather than an engine somebody could still
	// place orders with.
	automation, err := runInterlock(ctx, cfg.Engine.AutomationGate, paths, off, auditLog, clk.Now(), opts.Guardian)
	if err != nil {
		return nil, err
	}
	var guardian execgw.Guardian
	if automation.Verified {
		guardian = opts.Guardian
	}

	return &Context{
		Paths:       paths,
		Config:      cfg,
		TokenFile:   tokenFile,
		Official:    off,
		Automation:  automation,
		Guardian:    guardian,
		Audit:       auditLog,
		official:    off,
		broker:      broker,
		conditional: broker,
		// No lineage recorder: the engine records order lineage in the journal
		// transaction (design D2), not in the CLI's JSON cache. Until the journal
		// lands, trading.Service without a recorder behaves exactly as upstream's
		// does when lineage is unset.
		TradingService: trading.NewService(cfg.Trading, broker).WithConditional(broker),
	}, nil
}
