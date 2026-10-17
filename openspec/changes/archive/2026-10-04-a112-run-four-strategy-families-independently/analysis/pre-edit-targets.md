# Pre-edit target and impact matrix

This file records the current-base implementation targets discovered during proposal analysis. It is not permission to edit them. At apply time each `FLM-required` row must be refreshed against the actual base before its function body changes.

| Area | Existing symbol / file | Current constraint | Graph / risk impact | Pre-edit gate |
|---|---|---|---|---|
| M-B0 authority-origin citation | `Client.AuthorityOrigin`, `Client.authorityOriginLocked` in `internal/official/client.go` | production origin/transport proof exists, but default transport permits proxy and redirect behavior | new A112 seam may consume the opaque proof but must clone and narrow transport without changing these bodies | citation-only; any existing-body edit requires a new Manager amendment and full FLM/BTM/risk |
| M-B0 cached-token citation | `tokenManager.loadCache`, `isStillValid` in `internal/official/token.go` | ordinary `token()` may exchange/write and `refresh()` may exchange/adopt | new seam may perform its own locked owner/mode-checked read-only cache validation, but must not call token/refresh/exchange/save | citation-only; `token.go` remains read-only in M-B0 |
| M-B0 transport hazards | `Client.doRequest`, `Client.send`, `AttemptTrace`, `RateBudget` in `internal/official/{client,trace,ratebudget}.go` | ordinary path retries/refreshes and body/rate evidence is not same-request sealed for M-B | M-B0 must implement a separate single-attempt capped GET in the new leaf and preserve allow-listed raw headers | citation-only; every listed existing body remains read-only in M-B0 |
| M-B0 KR raw authority boundary | `Client.RawMinuteCandles`, `RawMinutePage`, `apiCandlePage` in `internal/official/{candle_raw,candle_reads}.go` | KR-only guard is intentional; decoded string cursor collapses absent/null/empty | US measurement must preserve raw cursor state without weakening KR production authority | citation-only; unchanged-US-rejection and tri-state-cursor REDs required |
| M-B0 new leaf | new `internal/official/a112_mbus_read.go` and tests | no current source; zero production callers | later one-shot tool only; any strategy/runtime/cmd caller would create circular authority | new-function AST/FLM/BTM/risk after RED/GREEN; static resolved-import/reference guard required |
| canonical registry | `strategyflow.ValidateDescriptors`, `canonicalDescriptor`, `LaneInput.matches` in `internal/strategyflow/registry.go` | exact six descriptors and six lane-kind map | Evaluate/Propose production integration and projection tests | FLM-required for each edited function |
| typed union/adapters | `internal/strategyflow/types.go`, `defaultRegistry`, `proposalRegistry` | six variants and positional descriptor bindings | all pure lane entry/proposal calls | AST/FLM for edited existing functions; exhaustive registry tests |
| route manifest | `validProductionRouteCandidates`, `productionRouteDescriptors` in `internal/strategyrouter/production.go` | exact three candidates per market; `len(seen)==3` | production scope loading and route tests | FLM-required |
| arbitration | `strategyrouter.Route` in `internal/strategyrouter/router.go` | active owner priority; raw int64 highest score; tie refusal | shared owner decision across every lane and campaign | citation-only if unchanged; FLM-required if score semantics change in body |
| scheduler capability typing | `AcquireRequest`, `Capability`, capability seal/token/fingerprint helpers in `internal/strategyrouter/quota.go` | market/horizon/poll scope without four-family contract | every low-priority strategy read admission and completion | FLM-required for every edited constructor/validator/seal helper |
| scheduler quota admission | `QuotaAuthority.Acquire`, `QuotaAuthority.Complete` in `internal/strategyrouter/quota.go` | one physical quota/commitment authority | all 8 workers, reset generation, issuance cap and safety reserve | high-risk FLM/BTM/risk required; concurrency/property RED first |
| proposal input | `buildLaneInput` (269-382), `validScopes` (468-480) in `internal/strategyproposal/production.go` | continuation/reversal branches, weekly fallback | `LoadProductionAuthorityBatch` and production tests | FLM-required; no generic fallback |
| production Route caller | `strategyRouteAuthorityLoader.collectMarket` / direct `strategyrouter.Route` call in `internal/app/engine/strategy_route_authority.go` | one pre-evaluation winner per market | downstream proposal loader, worker readiness and dispatch | L3 FLM-required; replace caller with sealed RouteSet without raw-score selection |
| production proposal caller | `strategyProposalAuthorityPair.ResultAuthority`, `strategyProposalAuthorityLoader.collectMarket` in `internal/app/engine/strategy_proposal_authority.go` | exactly one proposal per market | worker readiness, multi-symbol owner-scope intake | L3 FLM-required; remove market-wide singleton assumption without bypassing exact seal validation |
| worker readiness | `buildProductionStrategyMarketWorker` in `internal/app/engine/strategy_entry_supervisor.go` | one worker per market and exact one proposal | supervisor construction and dispatch-cycle tests | FLM-required |
| supervisor cardinality | `NewStrategyEntrySupervisor` in `internal/app/engine/strategy_entry_supervisor.go` | exact two market workers | runtime lifecycle, cancellation and worker ownership | L5 high-risk FLM/BTM/risk required before exact 8-worker/2-coordinator edit |
| production supervisor bootstrap | `Context.NewPairedStrategyEntryProductionAssembly`, `Context.NewRefreshingPairedStrategyEntrySupervisor` in `internal/app/engine/strategy_entry_supervisor.go` | assembles/refreshes exact two market workers and is called by `cmd/tossctl/engine.go` | production bootstrap, refresh and lifecycle ownership | L5 FLM-required unless a new explicit runtime replaces the call path; replacement requires citation/compile guard proving old bootstrap is unreachable |
| worker cycle | `Context.runProductionStrategyMarketCycle` | whole paired refresh, exact one proposal, direct shared dispatch | production worker, journal campaign CAS and dispatch | FLM-required |
| refresh boundary | `Context.refreshPairedStrategyEntryProductionAssembly` | one mutex around paired assembly and one-second cache | all market evaluation refresh | FLM-required if body changes; no remote I/O under lock |
| runtime projection read | `Context.Read` in `internal/app/engine/strategy_runtime_projection.go` | market-level snapshot | API/console/read RPC callers | L5 FLM-required; preserve snapshot consistency and legacy fields |
| runtime projection assembly | `strategyProjectionFromAssembly` in `internal/app/engine/strategy_runtime_projection.go` | two-market/one-lane projection | operator/API compatibility | L5 FLM-required; additive fixed-order lanes[8]/coordinators[2] only |
| evidence enum/typing | `kindSupportsMarket`, `validateTypedPayload` in `internal/strategyevidence/model.go` | exact kinds; shallow generic type checks | append/replay/projection/model tests | FLM-required for edited switch/parser; strict breakout parser preferred as new function |
| shared dispatch | `strategyDispatchCycle.dispatch` in `internal/app/engine/strategy_dispatch_cycle.go` | one protection→reconcile→q_final→lease→Gateway path | Guardian, journal, Gateway and broker mutation | reuse/citation target; edit only if lineage seal cannot be added at boundary, then full FLM |
| strategy risk authority | `strategyRiskAuthorityLoader.collectMarket` in `internal/app/engine/strategy_risk_authority.go` | shared q_final/risk validation | owner, Guardian and dispatch preconditions | L6 FLM-required only after all prerequisite gates; no second risk authority |
| account first-leg authority | `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority` in `internal/app/engine/strategy_account_first_leg_authority.go` | one account/owner admission | reservation, journal and dispatch | L6 high-risk FLM/BTM/risk required; first-leg-only and atomic owner invariant |
| first-leg admission bridge | `strategyFirstLegAdmissionBridge.admit`, `validateStrategyFirstLegResult`, `validateStrategyFirstLegAuthority` in `internal/app/engine/strategy_first_leg_admission.go` | validates result→authority→issuance boundary | q_final, owner, reservation and dispatch eligibility | L6 high-risk FLM/BTM/risk required; validation order and zero-broker refusal branches preserved |
| protection readiness | `productionProtectionAssemblies` (39-44), `execgw.ProfileProtection` constant (68) and `Gateway.checkProtection` (89-110) | both markets `Wired:false`; shipped scalar UNWIRED | startup interlock and every exposure-raising dispatch | prerequisite/citation only; no a112 weakening |

## Current-base bundle inventory

Frozen base is `016da6245feb60e13971388be386c2c2041469a8`. All source citations in this
matrix were re-read against that base. The auditable per-symbol
inventory is [`function-logic/inventory.json`](function-logic/inventory.json): it lists all
38 current-base function bundles with exact source range, SHA-256, AST branch/return/call
counts, bundle directory, and a non-claiming planned-RED disposition. It also records every
non-function/citation-only exception: `AcquireRequest`/`Capability` are types,
`execgw.ProfileProtection` is a constant, and the protection row remains a prerequisite
boundary. There are no extra or placeholder bundle directories.

Each current-base function bundle contains `ast.json`, `function-logic-map.md`,
`branch-test-map.md`, and `risk-pattern-report.md`. The owning lot must re-run CodeGraph
definition/callers/callees/impact and replace its planned RED rows with actual test evidence
immediately before changing any function body.

## Hard invariants for implementation review

- Descriptor/manifest/input/adapter expansion is atomic: no 7-lane or per-market 3/4 mixed authority.
- Exactly 8 evaluator instances and 2 coordinators may exist, but there is exactly 1 account-scoped Guardian/risk/dispatch/Gateway authority.
- Owner key remains `(account, market, symbol, position_generation)` and excludes family/horizon.
- Lane-local faults change only that lane's exposure-raising state; safety loop faults follow existing engine-safety policy.
- Existing a100 files and work-in-progress artifacts are outside this change and must not be edited or normalized by a112 work.
- No code, deploy step, migration or restart flips desired/effective lane, automation, autostart or LIVE approval.
