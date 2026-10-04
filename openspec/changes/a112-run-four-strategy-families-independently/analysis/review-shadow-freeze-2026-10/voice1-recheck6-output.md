# SHADOW re-freeze 재검 6라운드 voice1 — SubagentHandback 원문 (agent a8ef03efb4591dbc6, 저장 2026-10-05)

**Overall: PASS** (voice 1, re-check round 6, final — checking only that the five folds are reflected). All five are reflected. No P0. One implementation-stage RED candidate below.

Safety: I read nothing under `~/.codex` and did not open `_codex/`. I edited nothing in the repo. Every shell started with `set -euo pipefail`; this round needed only reads and greps. The v3.3 sha256 matches (df2d07bc…418ff). One harness note: the first Bash call was refused because a hook rewrote it and the auto-mode classifier gave no verdict. I reissued the identical call once and it ran; nothing else was affected.

| Fold | Reflected? | v3.3 coordinate | Basis |
|---|---|---|---|
| 1. P1-E: narrower continuity claim, intended gap tested separately | Yes | §5.1 "no-rejection control (v3.3 narrowing)" | Now says only that an observation of the same wave is never rejected by age alone. The UNOBSERVED gap between `record` and the next publish has its own test, and the option of accepting the previous wave's observation is recorded as rejected. |
| 2. N5: defer safety | Yes | §5 "success flag" and "nil-safe runtime access" bullets; pins (iv), (v) and the central-integrity identity test | Every element I asked for is in: the accessor takes `strategyLanesMu`, lock order `strategyLanesMu` → `runtime.mu`, the epoch map is initialised in the constructor, `invalidateShadow` has a nil guard, and failure is judged by `returnedNil` rather than `recover()`. The shape pins allow the guard and the lock. Pin (v) covers the first cycle after boot (refresh error, no runtime, identical error, no panic, `FirstSwallowedFailure` unchanged). The identity test covers both return and panic. One small doc gap that does not affect this verdict: the brief says the conversion of the `recoverMarketLanes`/`persistMarketLatches` lock windows was rejected, but no grep hit in v3.3 records that rejection. My item was optional and ungraded, so one line in §5 or §13 would do. |
| 3. Elapsed time over the deadline counts as failure | Yes | §5 "deadline elapsed = failure"; pin (iv) "late nil after deadline + δ" | When elapsed ≥ `MaximumStrategyCycleLimit`, a nil return still discards the observation and does not start the shadow step. The boundary test uses limit vs limit − 1ns. |
| 4. Collection helper receiver | Yes | §4 "receiver (v3.3)" | The helper is either a method on an addressable local declared in `coordinateMarketProposals` (`var shadow strategyShadowBatch`, no `&`, not a pointer field or parameter) or a pure function returning a new slice. The receiver expression is in the shape pin. |
| 5. Fact correction (panic latches the market only for effective workers) | Yes | §5 "discard immediately on failure" bullet | Effective workers: panic → `latchMarket`. refreshOnly workers (today's production): swallowed and keep running (supervisor :899-901, :922-951). The discard works in both cases. |

**P0:** none.

**Implementation-stage RED candidate (outside the design loop, for real code, mutation and the gate):**
- Two clocks decide whether one cycle failed. The closure's `returnedNil` uses elapsed < `MaximumStrategyCycleLimit`; the supervisor watchdog uses a `select` in `invokeBoundedStrategyCycle` (strategy_entry_supervisor.go:1057-1076). Near the limit they can disagree. The cycle returns nil just under the limit, so the closure counts it as success and starts the shadow step, while the watchdog's `Sleep` fires first, or Go's `select` picks it randomly when both are ready. The supervisor then records a deadline failure: abandoned/abnormal, which latches an effective worker and is swallowed for refreshOnly.
- The closure never ran its invalidation, so a SHADOW can be published for a wave the supervisor counted as failed. This is display-only and needs no fix in the design.
- RED to write: at the boundary with a fake clock, make the result and the deadline ready at the same time. Assert the supervisor's judgement and the closure's judgement agree, or that the projection is UNOBSERVED whenever the supervisor recorded an abandon.

**Repo unchanged (start = end):**
```
 M docs/ROADMAP.md
 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
HEAD 4d22d72663dcf3a42be928b3ecf3db817ffc6ac5
```
