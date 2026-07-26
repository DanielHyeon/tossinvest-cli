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
	"path/filepath"
	"time"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// ErrOfficialCredentialsRequired is returned by New when no usable official
// Open API credentials are available. It is a startup refusal, not a warning:
// there is no fallback broker for the engine.
var ErrOfficialCredentialsRequired = errors.New(
	"official Open API credentials are required to start the engine: run `tossctl openapi login` " +
		"or set TOSSCTL_OPENAPI_KEY and TOSSCTL_OPENAPI_SECRET (the engine never falls back to the web session)")

// ErrAccountUnresolved means startup could not learn which account the
// credentials belong to (task 7.1, design D8 step 1).
//
// It is a refusal and not a warning, with the gate on or off. Every durable
// record the engine writes is scoped by the account reference — the journal's
// intents, the risk reservations, the RECONCILE states — and the gateway refuses
// to be constructed without one. An engine that cannot name its account is an
// engine whose records cannot be attributed, which is worse than an engine that
// did not start.
var ErrAccountUnresolved = errors.New(
	"engine: the account could not be resolved from GET /api/v1/accounts, so nothing the engine " +
		"records could be attributed to it")

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
	// Logger receives the structured line the startup interlock emits (task
	// 7.5). Nil discards it — obs.Logger's methods are nil-safe — which is what
	// the tests that only assert on the audit trail want.
	Logger *obs.Logger

	// protectionOverride satisfies interlock clause 6. Unexported, therefore
	// test-only, and reached through export_test.go.
	//
	// Clause 6 is an unmet constant in this build by design (interlock.go
	// profileProtection), so with no seam the tests that are about the *other*
	// clauses — the acceptance audit line, the structured log of a verified
	// start — could not reach an accepted start at all and would quietly stop
	// testing anything. The seam is here rather than on the exported API for the
	// same reason journalFSProber is: a production caller must not be able to
	// claim a capability the build does not have.
	protectionOverride *ProtectionReadiness

	// journalFSProber overrides the filesystem probe the journal's durability
	// guard uses. Unexported, therefore test-only — the same arrangement
	// journal.Options.migrationOverride uses, and for the same reason: the guard
	// exists because fsync does not mean fsync on every mount, and a production
	// path that could stub it out has turned the durability contract into a
	// comment. Tests reach it through export_test.go, which is not in the binary.
	journalFSProber journal.FSProber
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

	// AccountRef is the account the credentials resolve to, read once at startup
	// and independent of the automation gate (task 7.1). It is the reference
	// every durable record is scoped by.
	AccountRef string

	// Journal is the engine's durable record of every intent, attempt, fill,
	// reservation and RECONCILE state (task 7.2).
	//
	// It is exported because it is a record, not a mutator: nothing on it can
	// reach the broker. Close closes it, and the engine owns it — a caller that
	// closes this handle behind the engine's back breaks the gateway.
	Journal *journal.Journal

	// Gateway is the engine's only order mutation surface (task 7.3).
	//
	// Unlike a trading.Service it is safe to hand out: every method on it demands
	// a GuardianDecision that was persisted before the call, journals the intent
	// before dispatch, and settles the attempt afterwards. "Submit without an
	// authorisation" is not an API it can express.
	Gateway *execgw.Gateway

	// Entry is the gate new exposure is checked against — staleness and latches.
	// Exported because an operator surface has to be able to see a latch and
	// clear it; it opens nothing on its own.
	Entry *execgw.EntryGate

	// Resolver settles IN_DOUBT attempts by observation. It writes to the
	// journal and reads from the broker; it has no mutation path.
	Resolver *execgw.Resolver

	// Reconcile owns the RECONCILE projection: it was restored from the journal
	// at startup and it is what a reconciliation loop observes into.
	Reconcile *reconcile.Tracker

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

	// tradingService applies the user's config policy and the confirm-token gate
	// on top of the broker. It is what the gateway wraps.
	//
	// It is unexported since task 7.4, and that is the seal engine-safety asks
	// for: "엔진 컨텍스트는 mutation 메서드를 가진 서비스 값을 외부에 노출해서는
	// 안 되며". The confirm token it demands is no protection at all against a
	// caller holding the service — PreviewPlace hands the token back, so the gate
	// is one line of local arithmetic away from being satisfied. Anything that
	// needs to mutate goes through Context.Gateway, which cannot be satisfied
	// without a GuardianDecision that was persisted before the call.
	//
	// A caller that legitimately owns its own decision issuance and its own
	// journal — `tossctl flatten-all` — builds its order path with NewOrderPath
	// instead of taking one out of an engine it did not assemble.
	tradingService *trading.Service
}

// ConditionalMutator is the conditional-order write surface. It is an alias of
// trading.ConditionalBroker (the gate and intent assembly live there, task 1.4),
// so the engine's adapter and the CLI's hybrid client are interchangeable at
// this seam.
type ConditionalMutator = trading.ConditionalBroker

// OrderPath is the engine profile's order stack without the engine: the config,
// the official client and the policy-enforcing trading service, and nothing else.
//
// # Who this is for
//
// Exactly one caller: `tossctl flatten-all`, which is its own decision issuer and
// owns its own journal, entry gate and gateway (engine-safety: "기존 소비자인
// flatten은 엔진 컨텍스트가 아니라 자체 배선으로 구성한다"). It used to reach into
// engine.Context for a trading.Service, which is precisely the exposure task 7.4
// removes.
//
// # Why this is not the same hole with a new name
//
// A Context is what an operator, a loop or a future strategy holds; handing a
// mutator out of it means every one of them can bypass the journal and the
// Guardian by accident. An OrderPath is what a caller has to *ask for* by name,
// in a constructor whose doc says who may call it and what they take on: the
// obligation to journal the intent, to persist a decision, and to settle the
// attempt. That is a deliberate act one reviewer can see in one line of diff, not
// a field that happens to be within reach.
//
// It performs no network call and opens no journal. The caller supplies both.
type OrderPath struct {
	Paths     config.Paths
	Config    config.File
	TokenFile string

	// Official is the read-only view of the broker connection, the same
	// interface Context.Official carries.
	Official OfficialReads
	// Trading is the mutation-capable service. Wrapping it in an
	// execgw.Gateway before use is the caller's obligation.
	Trading *trading.Service

	// official, broker and conditional are what NewContext needs to finish
	// assembling an engine; they stay unexported for the reason the Context's
	// equivalents do.
	official    *official.Client
	broker      trading.Broker
	conditional ConditionalMutator
}

// NewOrderPath resolves the config and credentials and builds the order stack.
//
// The refusals are the engine profile's, unchanged: no credentials means no order
// path, because the alternative is a caller quietly trading through the scraped
// web session. Only the gate-related fields of Options are ignored — there is no
// gate here to interlock.
func NewOrderPath(opts Options) (*OrderPath, error) {
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

	// No lineage recorder: order lineage is recorded in the journal transaction
	// (design D2), not in the CLI's JSON cache. trading.Service without a
	// recorder behaves exactly as upstream's does when lineage is unset.
	return &OrderPath{
		Paths:       paths,
		Config:      cfg,
		TokenFile:   tokenFile,
		Official:    off,
		Trading:     trading.NewService(cfg.Trading, broker).WithConditional(broker),
		official:    off,
		broker:      broker,
		conditional: broker,
	}, nil
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
//
// With config's automation gate on, four more refusals apply — no Guardian, zero
// limits, no/expired/mismatched capability attestation — all in interlock.go.
//
// Since task 7.1 one broker read happens on every start, gate or no gate: the
// account resolution below. It is the only network call a gate-off engine makes,
// and it is a read the engine needs regardless of what the gate says.
func New(opts Options) (*Context, error) {
	return NewContext(context.Background(), opts)
}

// NewContext is New with a caller-supplied context. The context bounds the
// startup account read and, with the gate on, the interlock's checks.
//
// The construction order is fixed by engine-safety ("엔진 프로필은 다음 순서로
// 구성한다"): account resolution → journal → gateway → interlock. Each step is a
// precondition for the next, and the interlock runs last precisely so that it can
// verify what the earlier steps produced instead of taking their success on
// trust.
func NewContext(ctx context.Context, opts Options) (*Context, error) {
	path, err := NewOrderPath(opts)
	if err != nil {
		return nil, err
	}
	paths, cfg, tokenFile := path.Paths, path.Config, path.TokenFile
	off := path.official

	auditLog, err := openAuditLog(opts)
	if err != nil {
		return nil, err
	}
	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}

	gate := cfg.Engine.AutomationGate
	status := newAutomationStatus(gate, paths, opts.protectionReadiness())

	// The audit trail is written before anything can refuse: an operator's
	// settings change is worth recording whether or not the engine then agrees to
	// start on it, and the refusal record that may follow is what ties the two
	// together (§0.5).
	if err := recordGateSettings(auditLog, gate, status.AttestationFile); err != nil {
		// A settings change we cannot record is a settings change nobody can
		// audit. Refusing here is the conservative direction: the engine does not
		// start, and no order is placed off the record.
		return nil, fmt.Errorf("engine: recording the automation-gate audit trail: %w", err)
	}

	// --- 1. account resolution (unconditional) ------------------------------
	//
	// This used to be step 5 of the gate's verification, which meant a gate-off
	// engine never performed it. It is here now because it is not a gate concern:
	// the journal, the gateway and the reconcile projection are all scoped by the
	// account, and comparing an attestation against it is only one of its uses.
	accountRef, err := resolveAccountRef(ctx, off)
	if err != nil {
		return nil, refuseStartup(auditLog, gate, fmt.Errorf("%w: %w", ErrAccountUnresolved, err))
	}
	status.AccountRef = accountRef

	// --- 2. the journal -----------------------------------------------------
	jrn, err := openEngineJournal(ctx, opts, clk)
	if err != nil {
		return nil, refuseStartup(auditLog, gate, err)
	}

	// --- 3. the gateway -----------------------------------------------------
	wiring, err := buildGateway(ctx, gatewayInputs{
		journal:    jrn,
		trading:    path.Trading,
		official:   off,
		accountRef: accountRef,
		clock:      clk,
	})
	if err != nil {
		_ = jrn.Close()
		return nil, refuseStartup(auditLog, gate, err)
	}

	// --- 4. the interlock ---------------------------------------------------
	//
	// It runs before the context is returned, so a refused gate produces no
	// engine at all rather than an engine somebody could still place orders with,
	// and it runs *after* construction so that "a gateway exists" is something it
	// verifies rather than assumes.
	automation, err := runInterlock(status, gate, auditLog, opts.Logger, gateFacts{
		guardian:  opts.Guardian,
		trading:   cfg.Trading,
		gateway:   wiring.gateway,
		transport: path.broker,
		now:       clk.Now(),
	})
	if err != nil {
		_ = jrn.Close()
		return nil, err
	}
	var guardian execgw.Guardian
	if automation.Verified {
		guardian = opts.Guardian
	}

	return &Context{
		Paths:          paths,
		Config:         cfg,
		TokenFile:      tokenFile,
		Official:       off,
		AccountRef:     accountRef,
		Journal:        jrn,
		Gateway:        wiring.gateway,
		Entry:          wiring.entry,
		Resolver:       wiring.resolver,
		Reconcile:      wiring.reconcile,
		Automation:     automation,
		Guardian:       guardian,
		Audit:          auditLog,
		official:       off,
		broker:         path.broker,
		conditional:    path.conditional,
		tradingService: path.Trading,
	}, nil
}

// Close releases what the engine profile owns. It is safe on a nil context and
// safe to call twice.
func (c *Context) Close() error {
	if c == nil || c.Journal == nil {
		return nil
	}
	j := c.Journal
	c.Journal = nil
	return j.Close()
}

// spentNonceRetention is how long a nonce consumption record is kept.
//
// The invariant is "retention ≥ the longest decision TTL" (engine-safety: "소비
// 기록의 보존 기간은 최대 결정 유효 시간 이상이어야 한다"), and thirty days clears
// every decision this build can issue by three orders of magnitude. The number
// is not a tuning knob: it is a floor with a wide margin, and a shorter one would
// let a decision outlive the record of its own consumption.
const spentNonceRetention = 30 * 24 * time.Hour

// openEngineJournal opens the journal the engine records against and sweeps the
// consumption records once (task 7.2, design D8 step 2).
//
// # What this makes a startup condition
//
// The journal's filesystem allowlist (ext4/xfs/btrfs) and its integrity check
// now gate engine startup. That is the intended inheritance of the P1 journal
// contract rather than a side effect: an engine whose durability guarantees are
// unverifiable is an engine whose record of what it did is unverifiable, and the
// whole execution contract is built on that record. An operator on tmpfs, NFS or
// a fuse mount gets a refusal that names the filesystem, not a silent downgrade.
//
// # Why the path follows --config-dir
//
// This is flatten's convention, kept: an explicit config directory means an
// isolated run, and the journal follows it so a test or a second profile cannot
// touch the real one. With no override the journal resolves its own default
// ($TOSSOS_DATA_DIR > $XDG_DATA_HOME/tossos > ~/.local/share), which is where the
// journal lives for a normally installed engine.
func openEngineJournal(ctx context.Context, opts Options, clk clock.Clock) (*journal.Journal, error) {
	path := ""
	if opts.ConfigDir != "" {
		path = filepath.Join(opts.ConfigDir, journal.DBFileName)
	}
	j, err := journal.Open(ctx, journal.Options{
		Path:     path,
		Clock:    clk,
		FSProber: opts.journalFSProber,
	})
	if err != nil {
		return nil, fmt.Errorf("engine: opening the journal: %w", err)
	}

	// The retention sweep. A restart is the one moment the engine is provably not
	// in the middle of a dispatch, and the sweep is once per start rather than on
	// a timer: a background goroutine deleting rows is a second writer with no
	// transaction discipline, and it keeps running after the engine has stopped
	// trading.
	//
	// The retention is widened to cover the longest TTL actually on disk so the
	// invariant can never be *violated by this call* — the alternative, refusing
	// to start because some old decision was issued with a long TTL, would turn a
	// housekeeping detail into an outage while deleting nothing either way.
	retention := spentNonceRetention
	if longest, err := j.MaxDecisionTTL(ctx); err == nil && longest > retention {
		retention = longest
	}
	if _, err := j.PruneSpentNonces(ctx, clk.Now(), retention); err != nil {
		_ = j.Close()
		return nil, fmt.Errorf("engine: sweeping spent decision nonces: %w", err)
	}
	return j, nil
}
