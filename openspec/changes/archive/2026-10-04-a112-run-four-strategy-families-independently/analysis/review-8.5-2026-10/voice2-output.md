# 8.5 voice 2 (equivalence) — SubagentHandback 원문 (agent a1210b2bf3fb7a75a, 저장 2026-10-04)

I did not read or search anything under ~/.codex. The real repository is unchanged (measured at the end of this report).

# Voice 2: acceptance-set equivalence (a112 8.5)

**Verdict: HOLD.** There is one P1 and no P0. For B1 (activation refusals) and A (breakout 1.2 counterfactual), the before and after commits give identical verdicts on every input I generated. For B2 (`collectMarket`), every input I generated also matched. But one more input does not: the author claims the gate's inputs, including `ctx`, are value-equal at the old and new call sites, and that claim is false. I reproduced a market that closed before the edit and opens after it.

## Method
- I made four isolated copies with `git archive <commit> | tar -x`: `0b441270^`, `0b441270`, `65444341^`, `65444341`. They live under `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/voice2/`.
- Outside the change itself, the Go trees of each pair are identical (I checked with `diff -rq`), so the shared test helpers are the same on both sides.
- I copied the same driver files from `voice2/drivers/` into both copies of each pair and ran them with a private GOCACHE and `GOFLAGS=-trimpath`. The raw outputs are in `voice2/out/`.
- To confirm the drivers can actually see a difference, I ran positive controls: mutations of the "after" copy via `-overlay`.

## Comparison tables

### B1 — `LoadProductionFamilyActivation` (driver `zz_voice2_activation_diff_test.go`)

There were 13,714 inputs:
- the base fixture, run for both KR and US;
- 116 single mutations;
- every pair of mutations, for both markets;
- 3,000 seeded random combinations of 3 to 5 mutations.

The mutations cover all ten config bindings, all eleven body bindings, the lifetime checks (unparsable time, ordering, the 24h limit at exactly 23h, at 23h + 1ns and at 25h), the descriptors (unknown lane, unknown lane plus family drift, family/horizon/version drift, invalid desired or effective state, effective ON with desired OFF, duplicate, missing, none, the other market's lanes, reversed order), the file (missing, mode 600 or 444, symlink, trailing newline or document, unknown field, duplicate key, empty, truncated, oversized, not JSON, stale pin), revoked, and the context (nil, canceled, deadline, canceled with an Undeclared cause).

| Input class | Before | After | Equal? |
|---|---|---|---|
| All 13,714 cases: verdict seen by the engine (Undeclared / other error / accept) | — | — | Yes, 0 differences |
| Primary sentinel | Unavailable 11,468 · Undeclared 609 · ctx.Canceled 603 · Accept 331 · ctx.Deadline 285 · Expired 275 · Revoked 143 | identical counts | Yes, 0 differences |
| Accepted activation contents (generation, market, actor, expiry, ProtectionReady floor, lease, Desired/Effective for all 8 lanes) | 331 accepts | identical | Yes |
| Full `errors.Is` chain | `Unavailable` | `Unavailable` + `ErrProductionRouteUnavailable` (1,541 read-fault cases) | Different, no verdict impact (see P2-1) |
| Any case matching two activation sentinels, or an error returned together with a Verified activation | 0 | 0 | Yes |

Positive controls, applied to the after copy:

| Mutation | Verdict differences found by my driver | Repo tests that fail |
|---|---|---|
| Drop the `lifetime over maximum` check | 100 | `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`, `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` |
| Disable the `known && family` check | 35 | `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| Drop the `actor` check | 68 | the named-field test and the bindings test |
| Read fault wrapped as Undeclared | 1,541 | 4 tests |
| Drop the config `market` check | 0 | only the named-field test (see P2-2) |

That answers brief question 4 for B1: if a check is left out of `failedFields`, the message-level named-field test catches it, and so do the behavioural tests, except for the config `market` term.

### B2 — `collectMarket` (driver `zz_voice2_collect_diff_test.go`, tag `tossos_testseams`)

The driver runs 12 gate modes against 22 scenarios, 264 cases in total.
- **Modes:** verified; rollback on Unavailable, Expired or `ctx` error; undeclared; verified but with an error; verified with the continuation lane latched; verified with the reversal lane latched; the real loader unpinned; the real loader pinned with no file; each of those with a canceled `ctx`.
- **Scenarios:** ready with 1, 2 or 3 lanes; route not ready, no entries, schedule not ready, no schedule activation; FX not ready or invalid read; loader with no account, no load function or a `.` config dir; bad public key; duplicate symbol; empty symbol; load failed; digest mismatch; production fault; no accepted scope; arbitration tie; tie plus FX failure.

| Input class | Before | After | Equal? |
|---|---|---|---|
| Ready, Reason, Routed / Proposed / Refused / Gated counts, gated outcomes, queue drops, arbitration refusal and detail, fault, digests, entry list | — | — | Yes, 0 of 264 differ |
| `dispatchHandoffs()` count and how many hand off, plus the legacy `dispatchHandoff().Single()` | closed markets hand off 0 | 0 | Yes |
| Activation carried by FX / internal / authority-invalid / production-fault closures (108 cases) | not verified, gen 0, gate loaded 0 times | verified gen 7 (verified modes) or zero (others), gate loaded once | Differs, as intended |
| ROUTE_NOT_READY, FAMILY_GATE_CLOSED, ARBITRATION_REFUSED, NO_ACCEPTED, READY | — | — | Yes, including the activation |
| **Context canceled after the proposal load returns** (driver `zz_voice2_ctxrace_test.go`) | **FAMILY_GATE_CLOSED**, 0 proposed, 2 refused, gated outcome DORMANT, 0 handed off | **READY**, 2 proposed, activation verified, **2 handed off** | **No (P1)** |

QUEUE_OVERFLOW, collision and unresolved selection were not run as inputs; they were checked by reading the code only. They come after the gate in both versions and use the same gate value.

### A — breakout `Evaluate` (driver `zz_voice2_breakout_diff_test.go`)

The driver generated 4,000 seeded random snapshots, mixing KR and US. Each has 1 to 25 bars after the opening range, with RVOL drawn from {0, 0.9, 1.0, 1.199999, 1.2, 1.3, 1.499999, 1.5, 1.999999, 2.0, 2.5}, wick from {0, 0.1, 0.35, 0.350001, 0.9}, and random close, high and VolumeExpanded. Each snapshot was evaluated four ways: fresh, reusing the same prior, corrected (last bar's revision + 1, sometimes an appended bar), and corrected again on top of that. That is 16,000 evaluations.

| Input class | Before | After | Equal? |
|---|---|---|---|
| phase, refusal, diagnostic, candidate, final, proposalID, seal, snapshotDigest, setupID, configDigest, resistance and range low, transitions, lineage, `validDecision` | — | — | Yes, 0 of 16,000 differ |
| RVOLAdmission, at2.0, at2.5 | — | — | Yes |
| RVOLAt1200000 | false | true in 674 fresh, 674 same-prior, 662 corrected and 662 corrected-again evaluations | Differs as intended. All differences are false→true, and an independent oracle (the spec sentence re-implemented) agrees on all 4,000 after-commit cases. The same oracle disagrees with the before commit 674 times. |

Fresh phase mix: FIRST_TOUCH 1,249 · RETEST_WAIT 920 · TIMED_OUT 611 · PROPOSED 562 · RANGE_LOCKED 424 · INVALIDATED 234.

Positive control: making B7 also set `firstTouch` produced 454 decision differences, so the driver does see decision changes.

No package outside `breakoutlane` reads `Provenance()` or the `RVOLAt*` flags. `strategyflow/breakout.go` reads only Phase, Refusal, Final, Candidate, SnapshotDigest and ConfigDigest.

## Findings

| Level | Finding | Reproduction | Result | Suggestion |
|---|---|---|---|---|
| **P1** | B2: the claim that `schedule`, `routes`, `observedAt` and `ctx` "do not change between the two sites" is false. `familyGateFor` depends on `ctx.Err()`, which `LoadProductionFamilyActivation` checks at entry and exit, and on the file contents at the moment it reads them. Computing it earlier makes the gate's snapshot older by the whole proposal-load time. `strategyproposal.LoadProductionAuthorityBatch` checks `ctx.Err()` only at entry, so a cancel that lands after its last I/O reaches the gate in the old order but not in the new one. The same mechanism lets a revocation that lands during the proposal load be missed for that cycle. | `zz_voice2_ctxrace_test.go`: `loader.load` calls `cancel()` and then returns a valid batch; the `loadActivation` stub returns `ctx.Err()` when canceled, otherwise verified. Run with `go test -tags tossos_testseams -run TestVoice2CancelDuringProposalLoad` in both copies. | Before: FAMILY_GATE_CLOSED with 0 handed off. After: READY, verified, 2 handed off. This is the permissive direction. I did not show an order going out: downstream, the account loader (`internal/strategyaccount/production.go`) does check `ctx.Err()`, which very likely closes the cycle — but that is safety by coincidence. | Keep the early computation for the carried activation. Before `coordinateMarketProposals`, either re-check `ctx.Err()` and treat a hit as rolled back, or recompute the gate and require the same generation. Add this case as a test and correct the "값 동등" (value-equality) claim in the logic map. |
| P2-1 | B1: read-fault errors now also match `ErrProductionRouteUnavailable` (the second `%w`). Nothing checks for that on this path today — the engine checks only Undeclared — but any future classifier would misattribute these errors. | Driver chain column. | 1,541 of 13,714 cases. | Format the inner error with `%v`, or document the double identity. |
| P2-2 | B1: the config `market` (`name == ""`) term never changes a verdict. A directory-read failure plus the body's `validMarket` check mask it. Only the message test pins it. This predates the edit. | Overlay `pc3_drop_cfg_market`. | 0 verdict differences in 13,714. | Pin it with a test that stubs the read, or accept and note it. |
| P2-3 | A: neither literal `1_200_000` (entry path and B7) is tied to the golden `rvol_counterfactual_ppm[0]` (`goldens/breakout-evidence-and-sizing-v1.json:60`). No test reads that golden field, so the code and golden can drift apart silently. | `grep -rln rvol_counterfactual internal --include=*_test.go` finds nothing. | No binding. | Add a test that reads the golden value and checks it against the threshold behaviour. |

## Attacks that survived
- **B1, `known &&` grouping, lifetime term order and unparsable times as zero values:** every pair of mutations, plus 3,000 random combinations, showed no verdict difference.
- **B1, `market` term:** the before code (`name == ""` plus the body's `body.Market != config.Market || !validMarket`) and the after code agree.
- **B1, Load B5/B6 split:** pin-mismatch and read-fault cases keep the same verdict.
- **B1, Undeclared leaking in through `%w … %w`:** `readProductionRouteFile` returns only `ErrProductionRouteUnavailable` or `*PathError`, and the json error is wrapped. A context canceled with Undeclared as its *cause* still returns `ctx.Canceled` and never matches Undeclared. Zero Undeclared+other combinations.
- **B2, closed market carrying a verified activation (brief question 1):** handed-off count is 0 in both versions. The account loader fails on `len(entries) == 0`. Dispatch needs a handoff. The lanes run with an empty input, which already happens in the before version for non-owning lanes in a READY market.
- **A, prior reuse and correction:** these return the stored decision, flags included, with seal and decision unchanged. B7 cannot overwrite the at1.2 value of an admitted bar, because the code breaks out of the loop on admission.

## Repo unchanged
`git -C /mnt/D/Axipient/workspace/TossOS status --short` and `rev-parse HEAD` were the same before and after (compared with `diff`). HEAD is `556c3c1f6cf9026edbaa658a166e929ddbedb1d7`, and the same 3 untracked entries were there from the start.

The `breakoutlane` "root" test failures (5) appear in both archived copies because they have no `.git`; they are not related to the change. The private GOCACHE in scratch is 484 MB and can be deleted.

