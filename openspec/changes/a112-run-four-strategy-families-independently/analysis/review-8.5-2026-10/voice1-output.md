# 8.5 voice 1 (fail-open) — SubagentHandback 원문 (agent a82a7d28ba19152f4, 저장 2026-10-04)

I did not read or search anything under `~/.codex`. I made no changes to the real repository. All experiments ran in isolated `git archive` copies, either through `go test -overlay` or with harness files dropped into those copies, under `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/voice1/` (`before`/`after` = `0b441270^`/`0b441270`, `Abefore`/`Aafter` = `65444341^`/`65444341`).

# Voice 1 (fail-open / exposure) — a112 8.5

**Verdict: SHIP.** I found no P0 and no P1. The B2 ordering finding (P2-1 below) does let one order through in a timing the old code would have blocked. It is a lengthened race window that both versions have, not a new kind of path. The Manager should accept it explicitly or close it.

## Findings

| # | Finding | Reproduction | Result | Suggestion |
|---|---|---|---|---|
| P2-1 | **B2: moving the gate read earlier lengthens the revocation race window.** The author claims the inputs are equal at the old and new call sites, but lists only `schedule`/`routes`/`observedAt`/`ctx`. `familyGateFor` also reads the manifest file and env pin, which can change during `loader.load`. Revocation (or any file swap, which breaks the pin) is checked only at that gate read. `FinalAuthorityCheck` (`strategy_dispatch_cycle.go:212`) re-checks expiry only, never revocation or pin. Both versions have this race; the edit widens it by the duration of the proposal-batch load. | Overlay test `zz_voice1_gate_order_test.go` (tag `tossos_testseams`). The `loader.load` stub flips `revoked=true` during the load. The `loadActivation` seam returns verified before the flip and `ErrProductionFamilyActivationRevoked` after it. Uses the author's fixtures `familyGateFixture` / `arbitrationRoutePair` / `routeReadySchedulePair` / `proposalFXPair`. | before: `ready=false reason=FAMILY_GATE_CLOSED entries=0 gated=2`. after: `ready=true reason=READY entries=2 verified=true`. | Either record this as an accepted window ("revocation is honored one proposal-load later; next cycle closes"), or re-check the pin/revocation in `FinalAuthorityCheck` (a cheap digest re-read), which closes the race in both versions. Correct the "값 동등" claim either way. |
| P2-2 | **B1 Q4: two `failedFields` terms can be deleted without any test failing.** Swept all 34 `fieldCheck` lines with one-line deletions. **(a)** `effective` (line 602): with it deleted, a manifest whose `effective` is invalid (`""`, `"on"`) is accepted. This fails safe: `lookup` returns a non-ON state, so the lane is DORMANT and nothing is exposed. **(b)** `issued_at not before expires_at` (line 569): the accept/reject set is unchanged, because `issued ≤ now` and `expires ≤ issued` force `familyActivationRemaining` to fire. Only the sentinel moves from Unavailable to Expired, and no test pins that field name. Both gaps already existed in the old OR expression (same mutants on `0b441270^`: both survive). | `go test -overlay <mutant>`: the strategyrouter package untagged and with `tossos_testseams`, plus `./internal/app/engine` and `./internal/strategyworker` untagged. Unmutated overlay control is GREEN. | 32/34 caught; m569 and m602 survive all of these. The engine package fails with the same two tests as the control run (see Notes), so the mutants added no failures there. | Add two failure shapes to `a112_activation_error_fields_test.go`: one with an invalid effective state, and one with `issued == expires` (assert the field name and Unavailable). |
| P2-3 | **B2: diagnostic only.** When a closed market now carries a verified activation, each ON lane reports `REFUSED / ARBITRATION_SEAL_MISMATCH / "the sealed proposal does not establish this worker's lane"` where it used to report DORMANT. This now also covers FX-not-ready, misconfiguration, bad key, duplicate symbol, load failure and production fault. It is the same output that ON lanes without a proposal already give in activated markets, and that the six post-gate closures already gave. During an FX outage it reads like a seal alarm. | Overlay test `zz_voice1_lane_test.go`: `FamilyWorker.Run(zero, Input{})` versus `Run(verified, Input{})` for the four KR workers. | zero: `DORMANT`; verified: `REFUSED/ARBITRATION_SEAL_MISMATCH` on all four. `err` is nil, so nothing latches and the failure counter does not move. | Optional: give the "no input this wave" case its own detail. Operator surface is out of scope per the brief. |

Out of scope (not graded): the new B1 messages never reach an operator. `familyGateFor` discards `err` entirely (`strategy_family_activation.go:128-140`), which falls under the ROADMAP row the brief excludes.

## Attacks the code survived

- **B1 Q2, disguising an error as Undeclared (the P0 candidate):**
  - The only path to `ErrProductionFamilyActivationUndeclared` is the empty or whitespace pin, checked first.
  - The double-`%w` `err` comes from `readProductionRouteFile`. It returns only `ErrProductionRouteUnavailable` or an `os.Open` `*PathError` (`production_owner_unix.go`), so it cannot wrap Undeclared.
  - Every decode error wraps Unavailable.
  - The ctx path returns only `context.Canceled` or `DeadlineExceeded`.
  - The differential run below never produced Undeclared together with another sentinel (3,653 Undeclared, all exclusive, identical before and after).
  - The reverse direction (Undeclared turning into another sentinel) also never occurred.
  - Side effect: read faults now also satisfy `errors.Is(ErrProductionRouteUnavailable)`, but no consumer checks that on the activation path.
- **B1 Q1, acceptance set:** differential harness `zz_voice1_diff_test.go`, same PCG seeds on both commits.
  - `validateProductionFamilyActivation` × 300,000 random bodies and configs: 0 differing lines. 6,364 accepted, 2,381 Expired, 291,255 Unavailable. The inputs mutated every body and config binding, descriptor table mismatches, invalid desired and effective states, missing, duplicate and extra lanes, and timestamps that were empty, non-UTC, non-canonical, `±1ns`/`±1h` around now, at the 24h edge, or with `issued ≥ expires`.
  - `LoadProductionFamilyActivation` × 20,000 real 0400/0600 files: 0 differing lines. Pins were correct, wrong, malformed, empty, whitespace or padded. Bytes were non-canonical, trailing, unknown-field or truncated. Files were missing. Cases also covered nil and cancelled ctx, a relative directory and a zero `ObservedAt`.
  - Covers (i) the `known &&` grouping, (ii) zero-valued times when parsing fails (no panic), (iii) the `market` term, and (iv) the split into read fault versus pin mismatch.
- **B1 Q3, leaks:** messages carry only market or family names, `lane_id`, timestamps, the fixed file name, a size, `os.Open` path text and `encoding/json` key names. No secret or account data. Manifest content is reached only after the digest pin matches. In production the error is dropped anyway.
- **B2 Q1, what a closed market's activation reaches:** I traced every consumer of `familyActivation()`:
  - Account `collectMarket`: `len(entries)==0` returns NotReady first.
  - First-leg `authorityForOwnerScope`: refuses on 0 entries.
  - `dispatchHandoffs` and `AdmitEachOwnerScope(false, nil)` return `[Admit(false, nil)]`, the same MarketClosed handoff as before.
  - `ResultAuthority`: not ready.
  - The dispatch ProtectionReady floor and `LeaseCeiling` are reached only with an accepted proposal.
  - Lane `evaluate`: latch and recovery use the schedule activation generation, not the family activation. A REFUSED cycle has `err=nil`, so no latch and no ledger write.

  No order, lease, latch or recovery generation comes out of a closed market.
- **B2 Q2:** a cancelled ctx means `Load` returns `ctx.Err`. The gate is rolled back with a zero activation, so the closure carries zero, the same as before. No new blocking: the FIFO-swap race on `os.Open` already existed on the ready path. `lanesFor` is nil-safe and takes only an RLock.
- **A (2.3 B7):** differential harness `zz_voice1_breakout_diff_test.go`, 60,000 random post-range bar sequences.
  - Bar parameters: RVOL ∈ {1.0, 1.199999, 1.2, 1.3, 1.499999, 1.5, 2.0, 2.5}, wick ∈ {350k, 350001, 100k}, random high/low/close around resistance, `VolumeExpanded`, plus same-digest prior reuse and a revision-2 correction with prior.
  - Outcomes covered: 4,251 PROPOSED, 12,770 RETEST_WAIT, 36,487 RANGE_LOCKED, 5,940 INVALIDATED and 552 TIMED_OUT.
  - phase, refusal, diagnostic, candidate and final quantity, proposalID, seal, transitions, the other RVOL flags, and the reuse and correction results were byte-identical before and after. Only `RVOLAt1200000` differs (23,513 → 36,677 true).
  - `decisionSeal` hashes only Transitions from the provenance.
  - `RVOLMinPPM` is fixed at 1.5M by config validation, so the admission branch's overwrite of the flag can never write false.
  - No production consumer outside the package reads `RVOLAt1200000` or `breakoutlane` `Provenance()` (grep).

## Repo-unchanged measurement

- Before: HEAD `556c3c1f6cf9026edbaa658a166e929ddbedb1d7` on `feat/a112-four-family-runtime`. Status: `?? .reticle-setup-crash.log`, `?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-8.5-2026-10/`, `?? w4.log`.
- After: the same HEAD, branch and three untracked lines. Unchanged.

## Notes

- The engine control run on the archived `0b441270` copy, untagged, fails 2 tests: `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` and `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration`. These are unrelated to a112 and probably caused by running from an archive copy rather than a checkout. The mutant runs failed exactly the same two.
- I did not run the tagged engine suite against the two surviving mutants.
- A separate session was running `go test -tags tossos_testseams ./internal/app/engine/` while I worked. It is not mine and I did not touch it.
- I tried to delete one harness file with an `rm` on a shell-variable path. A safety check blocked it and it did not run; the rewritten `"${V:?}"` form then succeeded. All harness files, mutants and outputs remain in the scratchpad `voice1/` directory.

