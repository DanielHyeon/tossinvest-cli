# SHADOW re-freeze 재검 voice1 — SubagentHandback 원문 (agent a8ef03efb4591dbc6, 저장 2026-10-04)

**Overall: FAIL.** Of my six first-round P1s, 2 are CLOSED (P1-2, P1-6) and 4 are PARTIAL (P1-1, P1-3, P1-4, P1-5). v2 also opens 2 new P1s: the carried shadow inputs lose their type at the accessor and ride inside the struct the order path reads (N1), and nothing pins that SHADOW clears inside the process (N2). No P0.

Safety: I read nothing under `~/.codex`. The review folder contains a repo-local `_codex/` directory; I did not open it. I edited nothing in the repo. Experiments ran only in a scratchpad copy of 2817064c (`…/scratchpad/v1copy2`), git archive → tar. Brief v2 sha256 matches (7f1e840c…37bd). Between c7219640 and 2817064c, `git diff --stat` shows no Go code changes, only docs and analysis files, so the code facts from round 1 still hold.

## First-round P1s

| ID | Verdict | v2 coordinate | Evidence / repro |
|---|---|---|---|
| P1-1 (SHADOW sees no input) | **PARTIAL** | §0 corrects the premise; §4 carries the pre-gate inputs | Undeclared markets are fixed: the pre-gate `batch.LanesFor` inputs reach the shadow. But §4 carries them **only on the success return**, and in a declared market a scope where only OFF families propose closes the whole market (`erasedScopes` → FAMILY_GATE_CLOSED). Copy repro `TestV1RProbe…` (tag `tossos_testseams`): partial ON with 005930 reversal (OFF) alone → `ready=false FAMILY_GATE_CLOSED`. Declared all-OFF, both families proposing → also `FAMILY_GATE_CLOSED`. Only partial ON with continuation and reversal together → READY. So in a declared market SHADOW observes an OFF lane only in scopes where an ON family also proposed, and never when every family is OFF. The §8 precondition "WOULD_EMIT ≥ 1 in a partial-ON declared market" only passes with a fixture where families co-propose. Fix: also carry the inputs on the FAMILY_GATE_CLOSED closure, or record the limitation explicitly and pin it. |
| P1-2 (FamilyShadow can mint ON) | **CLOSED** | §2 (separate package `internal/strategyshadow`, census ①, source freeze ④) | Copy repro: a sketch package outside strategyrouter does not compile. The twin-struct conversion gives `cannot convert twin{…} … to type strategyrouter.FamilyActivation`; the literal gives `cannot refer to unexported field generation`. Minor note, not graded: census ① counts only composite literals, so field assignment on a zero value inside strategyrouter (`a.generation = 1`) is not counted. That cannot happen from strategyshadow. To keep the boundary honest, also ban `unsafe`/`reflect` in strategyshadow's dependency closure. |
| P1-3 (no Envelope is not a boundary; zero-count pin is wrong) | **PARTIAL** | §8 differential; §2② and §4 reader census | The measurement is fixed: shadow pin on vs off, same Submit/handoff/gateway traces, two configurations, WOULD_EMIT ≥ 1 precondition. The reader boundary has a new hole, N1. |
| P1-4 (a shadow fault latches a lane or the market) | **PARTIAL** | §5 | Fixed: the shadow step runs outside `collect` and outside the bounded lane step, never returns an error, recovers its own panic, and has a fault pin. Open: no time budget is given. `a112_market_delivery_structure_test.go:8,171` pins the delivery call as the **last statement** of `runProductionStrategyMarketCycle`, so the shadow step must run before dispatch. That puts its latency on the order path, inside the market cycle budget `MaximumStrategyCycleLimit=30s` (strategy_entry_supervisor.go:30,362,401). Going over that budget means `ErrStrategyCycleDeadline`, an abnormal result, and a market latch. The lane `cycleDeadline` is also 30s (policy.go:85). The shadow does file I/O. Required: state the shadow's own deadline and the sum budget. The fault pin must inject a delay longer than that deadline and assert dispatch still runs in the same cycle and the market is not latched. Alternatively run the shadow step so it does not block dispatch. |
| P1-5 (sharing turns the activation loader into an edit target) | **PARTIAL** | §3 (activation loader not edited, copy pinned by AST on both sides, separate sentinels with exclusive `errors.Is`, cross-decode rejected both ways) | §3 says `readProductionRouteFile` is "already shared — each loader wraps it in its own sentinel". That does not hold across a package boundary: the function is unexported (production_owner_unix.go:12, production_owner_other.go:8), and so are `productionRouteOwnerUID`, `productionRouteDescriptors` (production.go:586), `productionRouteTime`/`Identity`/`DigestValid`/`Digest` (:751-768). §2 and §3 contradict each other: `strategyshadow` must either get exported wrappers from strategyrouter or copy the helpers, including the per-platform owner/0400 read and a **second descriptor table** (drift risk). The design is silent on which. Required: specify exported neutral wrappers in a new strategyrouter file without editing existing functions, a single descriptor-table source with a pin, and the helper layer added to the both-sides AST pin. |
| P1-6 (two separate "OFF lane" judgements) | **CLOSED** | §6 `ShadowEligible` = desired OFF ∧ effective OFF; `validateLane` uses the value form of the same rule; cases tested as a table; decision 63-v2 | Desired ON / effective OFF is excluded on the worker side, so the projection can no longer be rejected wholesale. The two encodings are pinned by the case table in §6. |

## New P1s opened by v2

**N1. The carried shadow inputs lose their type at the accessor and sit inside the struct the order path reads (v2 §4, §2②).**
- `strategyShadowInputs` is carried inside `strategyProposalMarketAuthority`. That struct is the receiver of the order-path methods:
  - `dispatchHandoffs` (strategy_dispatch_handoff.go:37)
  - `dispatchHandoff` (:18)
  - `authorityForOwnerScope` (strategy_first_leg_owner_scope.go:29) — the first-leg order authority seal.
- The §2② census counts functions that "see" the type through struct fields. Under that rule every method above "sees" it. The census is therefore either unsatisfiable or has to allowlist the dispatch methods. The planned structure pin covers `dispatchHandoffs` only, not `dispatchHandoff` or `authorityForOwnerScope`.
- More concretely: the "copy-returning accessor" hands back `[]strategyworker.Input`. That is the same type the lane runtime and `gate.admit` take (strategy_market_coordinator.go:91, strategy_lane_runtime.go:266). Past the accessor, the OFF-lane pre-gate inputs (Scope + SnapshotDigest + Proposal, enough to build an Envelope) are indistinguishable by type from admitted inputs, so a type census stops seeing them there.
- Fix:
  - Carry them in a separate value returned beside the authority, not inside it.
  - Have the accessor return a distinct opaque element type (e.g. `strategyworker.ShadowInput`) that only `ShadowVerdict` accepts, so `admit`/`Submit` cannot take it.
  - Count by field selectors and accessor call sites across the engine package, not by "functions whose types carry it".

**N2. No pin that SHADOW clears within a running process (v2 §5, §7, §10).**
- §5 says a failed load, panic or deadline leaves "no shadow observation for that wave", and §4 says a closure is "no observation". The §5 fault pin checks only latch, ledger, cycle error, legacy snapshot and market latch; it does not check the projection. §10 covers restart only.
- Missing pin: if the manifest is revoked, expires or fails binding mid-process, or the market closes, the next wave must project UNOBSERVED with `shadowOutcome` null. Otherwise an implementation that skips the record keeps showing the previous wave's SHADOW/WOULD_EMIT without a valid manifest, which breaks the :89 SHALL.
- Also pin that a shadow observation counts only when its wave equals the latest evaluate wave.

## First-round P2s

None got worse.

## Repo unchanged (start = end)

```
 M docs/ROADMAP.md
 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
HEAD 2817064cc1f78c4d0d6abc04211de906a7c3efc5
```

`docs/ROADMAP.md` was already modified when this re-check started, and nothing changed during it.
