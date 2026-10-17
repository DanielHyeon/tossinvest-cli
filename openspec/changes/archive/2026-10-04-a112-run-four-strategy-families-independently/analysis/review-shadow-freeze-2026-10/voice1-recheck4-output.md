# SHADOW re-freeze 재검 4라운드 voice1 — SubagentHandback 원문 (agent a8ef03efb4591dbc6, 저장 2026-10-05)

**Overall: FAIL** (voice 1, re-check round 4). N3 and the round-3 notes are CLOSED. N2 is PARTIAL: the remaining case is display-only, lasts at most until the manifest's 24h expiry, and needs one more condition to close. v3.1 opens one new P1, NP1: pin ① (A) fixes where the shadow code sits on the order path, but not what that code does. No P0.

Safety: I read nothing under `~/.codex`; the repo-local `_codex/` directory was not opened. I edited nothing in the repo. I reused my round-3 scratchpad copy of 4d22d726, which is still valid because the code has not changed. The v3.1 sha256 matches (c2ea7e82…671c).

| Item | Verdict | v3.1 coordinate | Evidence |
|---|---|---|---|
| N2 (stale SHADOW when no new wave comes) | **PARTIAL** | §5.1: expiry checked at projection time against `runtime.clk.Now()`, one function `shadowObservationUsable`, pin (e) with the `expiresAt−1ns` boundary | **Expiry is closed.** **Revocation while waves are stalled is still open.** If the cycle errors before `record` — a `recoverMarketLanes` error, or a lane panic re-raised at the join (strategy_lane_runtime.go:208-210, :244-247) — the wave counter does not move and the closure never starts a shadow step because it requires a nil return. The panic case also latches the market (`latchMarket`), which stops cycles (`evaluationState`, strategy_entry_supervisor.go:1035). The last SHADOW then still has an equal wave number and an unexpired manifest, so it stays visible until `expiresAt` (at most 24h) even after the manifest is revoked. That contradicts decision 63-v2: "no valid manifest (…not revoked…) → no SHADOW". Pin (a) tests revocation only when the next wave arrives. Cheapest fix: add an age bound to `shadowObservationUsable`, e.g. `now − observedAt ≤ k × PollInterval`. Clearing the market's shadow cell when the cycle closure gets a non-nil error would cover the error case but not the latched case, because no closure runs then. |
| N3 (where the shadow step gets its data; forbidden set; cross-market fault; race list) | **CLOSED** | §5 source: `evaluate` argument → `record` stores `{wave, batch}` in the same critical section that increments the wave → the closure copies it synchronously after a nil return; §2③ forbids the refresh, assembly and wave-join functions plus `strategyRefresh`/`strategyRefreshAt`/`strategyRefreshWave`, with a positive control; cross-market fault pin that also counts refresh leader calls; race list | The path where a shadow-led refresh wave fails and latches the other market is blocked. I checked how the cell lines up with `evaluate`'s error and panic branches (strategy_lane_runtime.go:199-258). An error before `record` leaves the cell and the wave at the old pair, and no shadow step starts. An error after `record` (`persistMarketLatches`, `staleLatchError`) leaves the cell at the new pair, but no shadow observation exists for that wave, so the projection shows UNOBSERVED. `record`'s early return when there are no observations (:323) moves neither. The two are always written together. If a cycle overrunning its deadline is abandoned and its closure later copies a newer cycle's cell, the copy is an atomic pair stamped with its own wave, so it stays consistent. Note: §5 copies "that market's activation", but the cell is `{wave, batch}`. State that the activation is stored in the same critical section; otherwise a re-read through `LoadProductionFamilyActivation`, which is not in the forbidden set, would evaluate eligibility against a different wave. |
| Round-3 notes (collision returns a partial batch; collection is one statement-level helper call) | **CLOSED** | §4: lineage collision → absent value (not an empty batch); AST pins ⓐ ⓑ ⓒ; absent vs empty batch behaviour test | Absent and empty can be told apart in a test: absent gives UNOBSERVED/null, an empty batch gives OFF lanes NO_INPUT. Pin ⓑ plus the behaviour test catches a mutation that returns the partial batch. |

## New P1 (opened by v3.1)

**NP1. Two shadow calls on the order path run before dispatch, and nothing pins their function bodies.**

These are the paths where pin ① (A) still lets shadow work in ahead of dispatch.

1. **The `evaluate` argument.** Pin ① allows the argument shape `fresh.<shadow field>.forMarket(market)`. That is a **method call on the shadow pair type** that runs before dispatch. The same sentence says "zero shadow function calls in the body", so the wording contradicts itself. ④ and ⑤ restrict only what `evaluate` and `record` do with the batch. They say nothing about `forMarket`'s own body, and census ③ covers only the async shadow stage. If an implementation puts the per-wave manifest load inside `forMarket`, that is file I/O before dispatch, and a panic there becomes a cycle panic, which latches the market. Fix: pin `strategyShadowPair.forMarket` by AST to the same shape as `strategyProposalAuthorityPair.forMarket` (strategy_proposal_authority.go:167-175: a switch, field returns, no calls). Rephrase pin ① as a type rule: in this function, the only shadow-typed expression is that one argument subtree, and its only call is that accessor.
2. **The collection helper in `coordinateMarketProposals` (§4 ⓐ).** It runs per proposal, before `gate.admit`. ⓐ fixes its position and that it is called once, not its body. A panic there would become INTERNAL_FAILURE through `collect`'s recover (strategy_proposal_authority.go:270-276), closing that wave's legacy dispatch. Fix: pin the helper body to a single `append` of the `strategyworker.ShadowInput` constructor's result, pin the constructor body to a composite literal, and add a fault pin: inject a panic into the collection helper → coordination and dispatch unchanged.

Minor wording, not graded: ④ says "no calls involving the batch", but `evaluate` has to call `record` with it. It should read "no method calls on or iteration over the batch; passing it to `record` and one stored assignment are allowed."

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
