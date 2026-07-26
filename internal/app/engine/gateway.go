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
	"fmt"
	"net/http"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
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
	journal    *journal.Journal
	trading    *trading.Service
	official   *official.Client
	accountRef string
	clock      clock.Clock
}

// engineWiring is the execution stack the engine profile owns.
type engineWiring struct {
	gateway   *execgw.Gateway
	entry     *execgw.EntryGate
	resolver  *execgw.Resolver
	reconcile *reconcile.Tracker
}

// buildGateway constructs the execution stack and restores its projections.
func buildGateway(ctx context.Context, in gatewayInputs) (engineWiring, error) {
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

	resolver := &execgw.Resolver{
		Journal: in.journal,
		Orders:  orders,
		Order:   orders,
		Account: account,
		Clock:   in.clock,
		Gate:    entry,
	}

	gateway, err := execgw.New(execgw.Options{
		Journal:    in.journal,
		Trading:    in.trading,
		Clock:      in.clock,
		AccountRef: in.accountRef,
		Source:     engineSource,
		Entry:      entry,
		Preflight:  &execgw.Preflight{Account: account, Clock: in.clock},
		Orders:     orders,
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

	return engineWiring{
		gateway:   gateway,
		entry:     entry,
		resolver:  resolver,
		reconcile: tracker,
	}, nil
}
