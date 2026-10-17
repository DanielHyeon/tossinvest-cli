# Current-main evidence

## Baseline

- Frozen base: `016da6245feb60e13971388be386c2c2041469a8` (`[A100 R0/M0] causal receipt measurement tooling (#3)`).
- **Landing re-freeze (2026-08-17):** the L0–L2, M-B, L1a and L1b lots were committed on the branch on top of `016da624`
  (commits `7d628041`, `5f73ea74`, `a78ee6ff`, `29f05054`, `5a44d424`), then `origin/main` (`a3671987`, 50 commits:
  a109 closure, a113–a115 registration) was merged as `d9d9f71f`. `base-commit.txt` now points at that merge commit
  so later lots (L1c, L3–L7) diff only their own edits. Proven before re-freezing: `git diff --name-only 016da624
  a3671987` touches no file under `internal/strategyevidence`, `internal/official`, `internal/scheduler`,
  `internal/clock`, `internal/breakoutlane`, `internal/officialbars` or `tools/logic-map`, so every FLM/AST bundle's
  `source_sha256` and the analysis above stay valid for the merged tree; `go test ./...` on the merge is green.
- **L1c re-freeze (2026-08-27):** `base-commit.txt` now points at `a8c3d067` (`fix(strategymarket,official): 분봉 시각표는
  봉이 닫힌 시각이다 [a117]`), the branch and `main` HEAD when the L1c lot opened. Reason: **a117** landed on `main`
  after `d9d9f71f`, so a112's diff window had swallowed another change's edits and `check_analysis` demanded bundles for
  `internal/strategymarket/bars.go:aggregateClosedKRXFiveMinute` and four of a117's test helpers. Proven before
  re-freezing: `git show --name-only` for the intervening commits lists only openspec/PM files plus
  `internal/official/candle_reads.go` (comment), `internal/strategymarket/*`, `internal/strategycandle/adapter_test.go`
  and `internal/strategyengine/lane_test.go` — none of them the `file` of any bundle under
  `analysis/function-logic/`, so every `source_sha256` in this change stays valid. After the re-freeze `check_analysis`
  is clean and later lots diff only their own edits.
- Repository: TossOS modular monolith; no new process, broker gateway or strategy database is required by this change.
- `codegraph sync .` current-base synchronization was re-run for L0. The per-symbol
  AST/FLM/BTM/risk inventory is `analysis/function-logic/inventory.json`; each body edit
  still requires a fresh CodeGraph definition/callers/callees/impact query immediately before
  the edit. CodeGraphContext/GBrain remain advisory only.

## A100 R0/M0 landing is not product ProtectionReady

The frozen commit includes the A100 R0/M0 causal measurement-tooling landing. It does **not**
complete A100 product work or mint protection readiness: current
`internal/app/engine/protection_wiring.go:39-44` constructs both KR and US assemblies with
`Wired:false`, and `internal/execgw/protection.go:64-68` fixes `ProfileProtection` at
`ProtectionUnwired`. Therefore all exposure-raising requests remain zero; build/offline/shadow
checks do not authorize container replacement or operating activation.

## Existing canonical strategy matrix

`internal/strategyflow/registry.go:12-19` declares exactly six descriptors:

| Family | KR | US | Horizon |
|---|---|---|---|
| continuation | `kr_short_flow_continuation_v1` | `us_short_participation_continuation_v1` | SHORT |
| reversal | `kr_short_absorption_reversal_v1` | `us_short_dislocation_reversal_v1` | SHORT |
| weekly-value | `kr_weekly_disclosure_value_v1` | `us_weekly_disclosure_value_v1` | WEEKLY |

Every descriptor is `Desired=OFF`, `Effective=OFF`, `Runtime=UNOBSERVED`. `ValidateDescriptors` at lines 25-41 requires the exact descriptor count and exact values. `canonicalDescriptor` at lines 109-117 and `LaneInput.matches` at lines 119-129 also enumerate the fixed set.

`internal/strategyflow/types.go:319-365` seals `LaneInput` as a six-variant tagged union. `internal/strategyflow/adapters.go:13-32` explicitly builds evaluation and proposal registries from the same six positions. A fourth family therefore requires coordinated registry, union, constructor and adapter changes; an unregistered lane cannot reach the evaluator.

## Production manifest and proposal assembly are exact-count gates

`internal/strategyrouter/production.go:373-412` requires `len(values)==len(want)`, validates each canonical descriptor and finally requires `len(seen)==3` for each market. The descriptor maps contain continuation, reversal and weekly-value only. Adding breakout to only one side is rejected rather than partially accepted.

`internal/strategyproposal/production.go:269-382` branches continuation → reversal → weekly-value when constructing typed lane input. Its `validScopes` production scope validator at lines 468-480 accepts exactly three lane IDs per market. Breakout needs an explicit, fail-closed construction branch and strict evidence decoding; a generic fallback to weekly-value is forbidden.

CodeGraph impact for `buildLaneInput` reaches `LoadProductionAuthorityBatch` and production integration tests. CodeGraph impact for `validProductionRouteCandidates` is centered on production route-scope loading and its exact-matrix tests. Those functions receive fresh Go AST/FLM/BTM/risk reports before edits.

## Current runtime is market-level, not lane-level

`internal/app/engine/strategy_entry_supervisor.go:377-417` builds one `StrategyMarketWorker` per KR/US market. Readiness requires all market authorities and `len(p.entries)==1`. `runProductionStrategyMarketCycle` at lines 419-440 refreshes the whole paired assembly, again accepts exactly one proposal, checks that the symbol campaign is FLAT/CLOSED and invokes the shared dispatch cycle.

`refreshPairedStrategyEntryProductionAssembly` at lines 443-459 protects the entire assembly with one mutex and a one-second cache. Remote or slow evidence work under this boundary would couple all lanes; implementation must collect upstream read-only evidence outside the shared critical section and publish immutable snapshots through bounded queues.

CodeGraph callers of `runProductionStrategyMarketCycle` are the refreshing supervisor and production market worker. Its dispatch callee is `strategyDispatchCycle.dispatch`; changing worker cardinality without a single coordinator would let multiple workers race toward the same shared authority.

## Shared arbitration and mutation boundary already exist

`internal/strategyrouter/router.go:195-292` uses owner key `(account, market, symbol, position_generation)`; horizon is intentionally not part of ownership. A current active owner wins. With no owner, the highest eligible ON candidate wins and an exact score tie returns `RefusalAmbiguous`. Because family scores are not currently defined on a common calibrated scale, this change must seal a versioned `score_ppm` contract or refuse cross-family comparison.

`internal/app/engine/strategy_dispatch_cycle.go:53-138` is the existing single mutation path. It validates the proposal, re-reads schedule/FX/protection/reconciliation, performs q_final admission, loads the Guardian decision, issues and claims a fenced dispatch lease, then calls the official Gateway. Lane workers must emit immutable proposals only and must not receive this gateway or journal writer.

## Evidence and protection constraints

`internal/strategyevidence/model.go:18-26` defines only tradability, disclosure and flow/participation evidence kinds. `kindSupportsMarket` at lines 217-227 is an exact switch. `validateTypedPayload` at lines 398-427 performs only shallow generic field typing. Breakout bars/state therefore require an additive kind and a dedicated strict decoder with integer minor/PPM units, bounded arrays and unknown-field refusal.

The SQLite evidence store is append-only and snapshot-based; its string kind column can accept an additive kind without a destructive schema rewrite. The producer must append only closed official bars, preserve revisions/corrections and seal point-in-time snapshots. Derived rolling windows are caches, not authority.

`internal/app/engine/protection_wiring.go:39-43` still constructs KR/US protection assemblies with `Wired:false`. `internal/execgw/protection.go:54-68` states that nothing in the shipped build produces `ProtectionWired`. Consequently this change cannot activate exposure-raising execution; a100 completion and a fresh market-scoped protection attestation remain hard prerequisites.

## Required change shape

1. Add pure `internal/breakoutlane` and strict evidence contracts.
2. Expand the canonical family matrix atomically from 6 to 8 descriptors.
3. Replace one-market/one-proposal evaluation with 8 independently supervised lane workers feeding two market coordinators.
4. Preserve one symbol owner and one account-wide risk/Guardian/dispatch domain.
5. Keep every new lane OFF and expose no activation minting surface.
6. Prove lane-local failure isolation while exit, fill, reconcile and protection supervision continue.
