# SHADOW re-freeze 재검 5라운드 voice1 — SubagentHandback 원문 (agent a8ef03efb4591dbc6, 저장 2026-10-05)

**Overall: FAIL** (voice 1, re-check round 5). All four of my remaining items are CLOSED. v3.2 opens one new P1, N5: the defer it adds can change the cycle's error or panic path. No P0.

Safety: I read nothing under `~/.codex`; the repo-local `_codex/` directory was not opened. I edited nothing in the repo. I reused my scratchpad copy of 4d22d726 (no Go changes since) and started every shell with `set -euo pipefail`. The v3.2 sha256 matches (68512bc3…09af1).

| Item | Verdict | v3.2 coordinate | Evidence |
|---|---|---|---|
| N2 (revocation while waves are stalled) | **CLOSED** | §5 "discard immediately on failure + epoch CAS" (pins i–iv); §5.1 age cap 74s, boundary pin, no-false-rejection control, pin (f) | Both paths from my round-4 finding are covered. **Failure path:** a failure before `record` (`recoverMarketLanes` error, panic re-raised at the lane join) makes the defer bump the epoch and delete the observation. A late in-flight result is then refused by the epoch-and-wave CAS. **Stall without failure** (supervisor stopped, not accepting, market latched with no cycle running): the age cap applies. **The 74s value is derived correctly.** `runMarket` (strategy_entry_supervisor.go:888-930) runs a cycle bounded by `cycleLimit=Maximum…=30s` (:362, :401), and the poll interval is `DefaultStrategyCycleLimit=5s` (:394). Add the 2s shadow deadline and the worst healthy gap between two publishes is 37s, so 2× gives 74s. 2×PollInterval (10s) would have flickered on healthy cycles. Revocation during a stall now shows for at most 74s, not 24h. |
| NP1-1 (`forMarket` body and the `evaluate` argument) | **CLOSED** | §5 pin ① as a type rule (only shadow-typed subtree is `evaluate` argument index 5; its only call is the accessor) + `forMarket` shape pin | The new argument is appended at index 5, which leaves the existing recovery-generation pin on `Args[2]` (a112_lane_latch_durability_test.go:335-348) untouched. The accessor is pinned to the same shape as its sibling (strategy_proposal_authority.go:167-175). |
| NP1-2 (collection helper before `admit`) | **CLOSED** | §4 helper and constructor shape pins; reason for not adding a panic-injection pin | The reason holds: one `append` of a composite literal of value fields has no nil dereference, index or conversion. Not graded, but it should be added to the shape pin: the helper must be called on an addressable local value declared inside `coordinateMarketProposals` (or be a pure function that returns the new slice). Called through a nil `*strategyShadowBatch` field or parameter, the same `append` dereferences nil and panics. That panic would become INTERNAL_FAILURE through `collect`'s recover (strategy_proposal_authority.go:270-276). |
| Note: where the shadow step gets its activation | **CLOSED** | §5 cell `{wave, batch, activation}` written in one critical section; §2③ forbids `loadFamilyActivation`, `familyGateFor` and `LoadProductionFamilyActivation` | Eligibility now uses the same wave's `promotion` (strategy_lane_runtime.go:199-201). |

## New P1

**N5. The §5 defer shape pin does not provide for getting the lane runtime safely, so the defer can rewrite the cycle's error or panic.**
- The pin says the defer body is "one call to `invalidateShadow(market)`" and the method body is "lock, increment epoch, delete observation only".
- The closure has to get the runtime from `c.strategyLanes`. That field is written under `strategyLanesMu` (strategy_lane_runtime.go:118-128), and `Read` also takes that mutex to read it (strategy_runtime_projection.go:49-51). A one-call defer can only read the field without the lock, which is a data race on a field set during the first cycle. Taking the lock breaks the one-call shape.
- **The runtime can be nil.** `runProductionStrategyMarketCycle` runs the refresh **before** `productionStrategyLanes` (strategy_entry_supervisor.go:515-518 before :539). If the refresh has failed every cycle since boot, `c.strategyLanes` is still nil. The defer then calls `invalidateShadow` on a nil receiver; reaching `runtime.mu` dereferences nil and panics.
- That panic replaces the original error, and the cycle path changes:
  - In refresh-only mode, which is production today, the error is swallowed and counted. The operator's `FirstSwallowedFailure` then shows a nil-pointer panic instead of the real refresh cause (`recordSwallowedCycleError`, :984-996).
  - In effective mode the latch refusal changes from Failure to Abnormal (`latchMarket`, :1098-1101).
  - Any cycle error carrying central integrity (`isCentralStrategyIntegrity`, :167-170, used at :950 and :957) would lose that identity, so `blockEntryOnCentralIntegrity` would not close new entries. `StrategyCentralIntegrityFailure` has no producer in the production refresh path today (grep), so this part is latent, but the defer sits exactly on the path that classification depends on.
- Fix:
  - Get the runtime through one nil-safe accessor that takes `strategyLanesMu`. The lock order `strategyLanesMu` → `runtime.mu` is the same as `productionStrategyLanes`, so there is no inversion.
  - `invalidateShadow` gets a nil-receiver guard (the runtime's existing convention), and the epoch map is initialised in the constructor (assigning into a nil map panics).
  - Allow both in the shape pin.
  - Add pin (v): refresh error on the first cycle with no lane runtime → the returned error is identical (`errors.Is` and same message), no panic, refresh-only `FirstSwallowedFailure` unchanged.
  - Add a passthrough test with the production closure: a central-integrity error returned by or panicking from the cycle keeps its identity through the defer.
- Non-graded note: `recoverMarketLanes` (strategy_lane_latch.go:156-169) and `persistMarketLatches` (:259-261) lock and unlock without `defer`. A panic inside those windows leaves `runtime.mu` locked, so the new defer would block forever while the panic unwinds. The cycle would hang until the deadline instead of returning Abnormal. Any other `runtime.mu` user would block on that mutex anyway, so this adds no new harm. Optionally convert those two windows to `defer` unlock.

## Lock order (surviving attack)

The CAS publish and `invalidateShadow` touch only runtime-level state (epoch, cell, observations) under `runtime.mu`, with no lane method calls. Projection holds `runtime.mu` and then the lane's lock (`lane.Status`, strategy_lane_projection.go:35-49). The CAS takes no lane lock, so the two lock orders cannot conflict.

## Repo unchanged (start = end)

```
 M docs/ROADMAP.md
 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
HEAD 4d22d72663dcf3a42be928b3ecf3db817ffc6ac5
```
