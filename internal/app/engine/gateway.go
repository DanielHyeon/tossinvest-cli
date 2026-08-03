package engine

// gateway.go assembles the engine's ExecutionGateway (extend-execution-contract
// task 7.3, design D8 step 3).
//
// # Why this was missing
//
// `execgw.New` had exactly one caller before this file: the flatten CLI. So the
// engine profile — the thing the gateway exists for — had no gateway at all, and
// nothing said so out loud. engine-safety's "Gateway 없이 mutation을 낼 수 있는
// 엔진 구성은 존재해서는 안 된다" was true only because the engine had no order
// loop yet, which is not the same as being structurally true.
//
// # What "optional" means here
//
// Every dependency below is nil-able in execgw.Options, and each one is optional
// only in the sense that the narrower unit tests construct gateways without it.
// A production gateway that omits one silently drops a guarantee:
//
//	Orders     → a created order's identifier is never read back, so a place is
//	             confirmed on the broker's ack alone (issues.md 2026-07-26)
//	Entry      → nothing stops a new entry during a RECONCILE or a stale read
//	Preflight  → a nil preflight is not a skipped check, it is no check
//	Replay     → an IN_DOUBT attempt can only ever be resolved by observation
//
// The startup interlock asks execgw.Gateway.Wiring() rather than trusting this
// file to have been written correctly (task 7.5).
//
// # Replay is wired and switched off
//
// The transport is constructed, and the attestation predicate is deliberately
// nil, which is what makes the replay path inert: the entry point refuses with
// `replay_not_attested` before it can reach the transport [미측정 — 2b 전 비활성].
// Wiring the transport now means the thing 2b has to turn on is a predicate and
// not a pile of plumbing, and it means the dark path is compiled and reviewed
// rather than written under pressure later.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// engineSource is what the gateway records as the producer of its intents. It is
// not "flatten-all" and not a strategy id: the automated engine is one producer,
// and the journal's Source column is how an operator tells them apart.
const engineSource = "engine"

// replayTimeout bounds one replay request.
//
// The replay transport carries its own HTTP client rather than borrowing the
// official one, because a replay with no timeout is a replay that can outlive the
// idempotency key it is using (the documented validity is 10 minutes — openapi),
// and because the official client's 401-refresh retry is exactly wrong here: the
// replay is counted in the journal before it is sent, so a transport-level retry
// would send an uncounted request.
const replayTimeout = 15 * time.Second

// gatewayInputs are what the gateway is built from. Everything here already
// exists by the time it is called: the account is resolved, the journal is open.
type gatewayInputs struct {
	journal     *journal.Journal
	trading     *trading.Service
	official    *official.Client
	accountRef  string
	clock       clock.Clock
	configDir   string
	manifestPin string
	// logger and publisher are the alert path's two halves. Both may be nil; see
	// newNotifier for what a nil publisher costs.
	logger    *obs.Logger
	publisher obs.Publisher
	// protectionOverride is retained only for the engine startup-status test
	// seam. It is never forwarded to the execution gateway: scalar status cannot
	// authorize an exposure-raising mutation.
	protectionOverride *ProtectionReadiness
}

// engineWiring is the execution stack the engine profile owns.
type engineWiring struct {
	gateway   *execgw.Gateway
	entry     *execgw.EntryGate
	resolver  *execgw.Resolver
	reconcile *reconcile.Tracker
	// retrier and notifier are the two producers task 3.2 wired and nothing
	// constructed (issues.md: "엔진 프로필에는 아직 Retrier도 Notifier도 구성하는
	// 곳이 없다"). They are built here, with their escalation fields populated.
	retrier  *execgw.Retrier
	notifier *obs.Notifier
	// converge issues the adjustment that makes an ADJUSTMENT_APPLIED release
	// reachable for a quantity mismatch (task 7.5 succession).
	converge *reconcile.Converger
	// ingest folds external holdings in and alerts on them.
	ingest *reconcile.Ingestor
	// floor is the exit loop's RECONCILE cap source.
	floor *reconcileFloor
}

// ErrProjectionUnbound means the fill apply hook that produces the position
// projection was never bound (add-core-domain task 6.1/6.3 succession).
//
// It is a startup refusal because the failure it prevents is silent and opens
// the gate. Reconciliation's local state is the projection; with no Project hook
// the projection stays empty however many fills arrive, the engine's belief is
// "we hold nothing", and every holding the broker reports is then classified as
// an *external* position — which is a notice, not a block. The disagreement that
// should have stopped new entries reads as a fact about the world instead.
var ErrProjectionUnbound = errors.New(
	"engine: the fill apply hooks are not bound, so the position projection would stay empty and " +
		"reconciliation would believe the account holds nothing — every broker holding would read " +
		"as an external position instead of a disagreement")

// ErrExitApplierUnbound is the exit half of the same guard.
var ErrExitApplierUnbound = errors.New(
	"engine: the exit apply hook is not bound, so no fill would ever resolve a pending exit proposal — " +
		"the first take-profit a position proposes would stay outstanding forever and suppress every " +
		"proposal after it")

// bindApplyHooks connects the position projection to the fill transaction.
//
// Both writes have to be one commit (design D7's 원자 apply hook): a fill the
// projection never saw is a fill the reconciliation, the exit policy and the
// entry gate all disagree about. The journal owns the atomic point and refuses
// to own the domain rule, so the rule is injected here — once, at wiring time.
//
// Task 7.3 filled in the Exit half, by extending this literal exactly as the
// note that used to stand here predicted: SetApplyHooks refuses a second call,
// so there was no other place to put it, which is what makes "one wiring point"
// a property and not a habit.
//
// What the exit applier adds to the same transaction: the cumulative taken
// fraction a sale moved, the resolution of the proposal that sale answered, and
// the completion of a state whose position has closed. It decides nothing about
// protection — a fill is not an observation — and the handle it is given cannot
// write a baseline, a watermark or a level at all.
func bindApplyHooks(j *journal.Journal) error {
	if err := j.SetApplyHooks(journal.ApplyHooks{
		Project: journal.ProjectPosition,
		Exit:    journal.ApplyExitFill,
		// Task 8.1 extended the same literal again, for the same reason: the
		// frozen trade outcome is priced with the shared cost model, and the
		// journal must not own the operator's numbers. The default set is the
		// conservative `[미검증]` placeholder task 1.1 shipped for exactly this —
		// an unconfigured model would leave every close without an outcome row.
		Costs: costs.DefaultModel(),
	}); err != nil {
		return fmt.Errorf("engine: binding the fill apply hooks: %w", err)
	}
	return nil
}

// checkProjectionWired is the guard that makes the binding above load-bearing
// rather than customary.
//
// It is asked of the journal the profile is about to reconcile with, after
// construction, for the same reason the interlock asks execgw.Gateway.Wiring()
// instead of trusting gateway.go to have been written correctly: a future
// reordering that drops the binding would otherwise produce an engine that
// starts, trades, and is wrong about what it holds.
//
// It asks about the exit applier too (task 7.3). An unbound one is quieter than
// an unbound projection and not gate-opening — no proposal is ever submitted
// that was not judged — but it is not harmless either: nothing would resolve a
// pending proposal, so the first take-profit a position proposes would be the
// last thing its policy ever did. The two hooks are bound by one call, so a
// journal with one and not the other is a wiring bug rather than a
// configuration, and it is named as one.
func checkProjectionWired(j *journal.Journal) error {
	if j == nil {
		return errors.New("engine: no journal")
	}
	if !j.ProjectionBound() {
		return ErrProjectionUnbound
	}
	if !j.ExitApplierBound() {
		return ErrExitApplierUnbound
	}
	return nil
}

// buildGateway constructs the execution stack and restores its projections.
func buildGateway(ctx context.Context, in gatewayInputs) (engineWiring, error) {
	// Before anything that reads a belief out of the journal: the belief has a
	// producer. The RECONCILE projection restored below is compared against the
	// position projection, so an unbound hook here is a comparison against a
	// blank.
	if err := checkProjectionWired(in.journal); err != nil {
		return engineWiring{}, err
	}

	orders := execgw.OfficialOrders{Client: in.official}
	account := execgw.OfficialAccount{Client: in.official}

	// The entry gate takes the default staleness thresholds: a required query
	// that has never been observed blocks new entries, which is the correct state
	// for an engine that has not started polling yet.
	entry := execgw.NewEntryGate(in.clock, nil)

	// The RECONCILE projection. The journal is authoritative (design D6) and both
	// the tracker's block set and the gate's reconcile-family latches are rebuilt
	// from it here — without this call a restart silently clears every block a
	// disagreement raised, which is the failure persisting the states removed.
	tracker := &reconcile.Tracker{
		Clock:      in.clock,
		Gate:       entry,
		Journal:    in.journal,
		AccountRef: in.accountRef,
	}
	if err := tracker.Restore(ctx); err != nil {
		return engineWiring{}, fmt.Errorf("engine: restoring the RECONCILE projection: %w", err)
	}
	entry.SetAuthorityRefresh(func() error {
		return tracker.Refresh(context.Background())
	})

	resolver := &execgw.Resolver{
		Journal: in.journal,
		Orders:  orders,
		Order:   orders,
		Account: account,
		Clock:   in.clock,
		Gate:    entry,
	}

	buildDigest, toolDigest := productionProtectionDigests()
	assemblies := productionProtectionAssemblies(buildDigest)
	provider := protectionreadiness.NewProductionProvider(protectionreadiness.ProductionConfig{
		ConfigDir: in.configDir, AccountID: in.accountRef, ProfileID: "production",
		BuildDigest: buildDigest, ToolDigest: toolDigest, ManifestDigest: in.manifestPin,
		Now: in.clock.Now, SupervisorAssemblies: assemblies,
	})
	readiness, err := protection.NewPairedReadinessAdapter(provider, in.accountRef, "production", provider.RuntimeContracts())
	if err != nil {
		return engineWiring{}, fmt.Errorf("engine: constructing the paired read-only protection readiness adapter: %w", err)
	}
	gateway, err := execgw.New(execgw.Options{
		Journal:    in.journal,
		Trading:    in.trading,
		Clock:      in.clock,
		AccountRef: in.accountRef,
		Source:     engineSource,
		Entry:      entry,

		ProtectionReadiness: readiness,
		Preflight:           &execgw.Preflight{Account: account, Clock: in.clock},
		Orders:              orders,
		Replay: execgw.HTTPReplay{
			BaseURL: in.official.BaseURL(),
			HTTP:    &http.Client{Timeout: replayTimeout},
			Headers: in.official.AuthHeaders,
		},
		// Attested stays nil: the replay capability has not been verified against
		// the real broker, and nil is what makes the entry point resend nothing.
		// Turning replay on is 2b's job and it is one field.
		Attested: nil,
	})
	if err != nil {
		return engineWiring{}, fmt.Errorf("engine: constructing the execution gateway: %w", err)
	}

	// The alert path and the query retrier, in that order: the Retrier announces
	// its transitions through the Notifier, so the Notifier has to exist first.
	notifier := newNotifier(in.journal, entry, in.accountRef, in.logger, in.publisher, in.clock)
	retrier := newRetrier(in.journal, entry, in.accountRef, notifier, nil, in.clock)

	// The two writers a reconciliation pass needs. They are separate because the
	// events are: a holding with no local instance is folded in and reported
	// (Ingestor), and a quantity the account disagrees with is converged to the
	// account's own number (Converger). Both write, which is what makes the
	// ADJUSTMENT_APPLIED release reachable at all (issues.md, task 6.3).
	alerter := notifierAlerter{notifier: notifier}
	ingest := &reconcile.Ingestor{
		Journal: in.journal, Alert: alerter, AccountRef: in.accountRef,
	}
	converge := &reconcile.Converger{
		Journal: in.journal, Credit: tracker, Alert: alerter, AccountRef: in.accountRef,
	}

	return engineWiring{
		gateway:   gateway,
		entry:     entry,
		resolver:  resolver,
		reconcile: tracker,
		retrier:   retrier,
		notifier:  notifier,
		ingest:    ingest,
		converge:  converge,
		floor: &reconcileFloor{
			official: in.official, tracker: tracker, journal: in.journal,
			retrier: retrier, clk: in.clock, account: in.accountRef,
		},
	}, nil
}
