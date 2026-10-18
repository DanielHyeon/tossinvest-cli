# a149 gstack implementation review

**PASS — scoped code/security review and focused QA evidence review.** No unresolved actionable findings in final reviewed snapshot. Aggregate gate, Manager final acceptance and lifecycle completion remain separate requirements. This review establishes neither authoritative calendar/data verification nor runtime indicator integration.

Reviewed 2026-10-18, base `0551b8781d7141417b3c35db9c3eb345689c32cd`, worktree `/tmp/tossos-a148-impl-20261018`. Reviewer is a separate Codex-host agent, independent of Terra implementation writer. Reused previously read gstack review skill, complete two-pass checklist and security/testing specialist checklists. No Claude execution or cross-model review claimed. Separate adversarial review is managed independently and is not asserted by this report.

## Scope and workflow

Read approved a149 design/spec/tasks and Manager implementation-boundary clarification, pre-edit evidence, complete new calculator/tests/consumer documentation and tracked diff. Read existing exactDecimal implementation and local Go big.Rat parser implementation to validate numeric-bound behavior. Searched new public/state symbols across all Go files: only new calculator/tests consume them. Existing runtime production files, frozen fixture files and parsing helpers have no diff against base.

Applied critical checks for data safety, concurrency, untrusted inputs, injection and enum consumers, followed by numerical/coercion, time-window, completeness, documentation and test checks. No network, broker, SQL, subprocess, authentication, mutable shared accumulator or runtime caller is introduced. CalendarVerified and official source labels are explicitly caller assertions; provenance states unverified. New calculation never mints an official seal. Inputs are read and copied for canonicalization, not modified.

Only this review report was written by reviewer. No code edits, test/gate runs, operational actions, telemetry, upgrades, configuration or memory writes. No further agents dispatched. Skill fix-first actions were sent to assigned writer; onboarding/logging/fetch/cross-model procedures were excluded by assignment. Pinned local base used. Existing-function FLM is not applicable because existing functions are not edited; this report is source review, not a new AST-evidence claim.

## Resolved findings and revisions

**[P2, confidence 10/10] Short binary exponents bypassed numeric size bound.** Initial `session_features.go:97` used `if len(s) > 64` before `exactDecimal(s)`. Existing `bars.go:231` rejects `eE/` but big.Rat.SetString accepts p/P binary exponents. Local Go `math/big/ratconv.go:195–200` permits binary exponent magnitude up to 10,000,000 and shifts a big integer. Thus a short `1p10000000` input defeats the intended bounded decimal arithmetic, expanding allocation/work substantially across a session. No large payload was executed. Writer added leaf-local boundedFeatureDecimal, accepting only bounded digits, one dot and an optional initial sign before the existing parser; p/P, radix and underscore forms are refused. Final tests exercise binary exponent, hexadecimal and underscore rejection. Existing parser/runtime unaffected. Resolved by full reread.

Manager's overall readiness correction also verified: known now requires VWAP, EMA9 and slope all known; otherwise usable individual values remain labeled separately. Tests enforce overall not_ready for incomplete prefixes, warmup and zero volume. Independent-review-requested unequal HLC/volume and next-session fixtures are present. Consumer documentation describes final semantics.

## Contract and test evidence

- Numerical oracle: ten constant-price five-minute groups 101 through 110, equal volume, yield session VWAP 211/2; first-nine-close seed 105; next EMA 106; slope `(105.5/105−1)*100=10/21`. At 09:46, 09:40 and 09:45 boundary VWAPs 104.5 and 105 yield `100/209`, not the 09:46 value. Final tests assert values and boundary timestamps.
- Unequal fixture independently yields first TP 101 and second TP 317/3, respective total volumes 5 and 10: weighted VWAP `(505+3170/3)/15=937/9`, slope `2800/909`. This detects close-only or unweighted substitution. New trading-date fixture resets EMA warmup.
- Readiness/bounds tests cover open/first-minute/first-five labels, 44/45-minute warmup and 390-minute session; missing opening/middle/latest minute, missing five-minute prefix, total zero volume, missing evidence and unsupported US. Values distinguish genuine zero slope from unavailable values.
- Integrity tests cover future close/receipt, receipt before close, duplicate/reverse timestamps, mixed symbol/currency/date, adjusted basis, malformed OHLC, negative volume, numeric grammar/size and inconsistent five-minute OHLCV. Five-minute values are checked against their constituent minute OHLCV when present; incomplete minute prefix cannot produce ready EMA.
- Canonical digest copies input slices and normalizes numeric values/time instants; tests check replay equality, unchanged inputs and equivalent decimal/time-zone representation. Digest is input evidence, not an authentication seal.

Read writer-produced `/tmp/tossos-a149-focused.log`: strategymarket `ok` (0.513s), strategycandle `ok` (cached), strategyengine `ok` (0.907s), for writer's focused package suite. Reviewer did not independently run tests. These results do not prove aggregate-gate or whole-repository success.

Exact rational operations and bounded decimal fields/bar counts avoid floating-point overflow/rounding inside calculation. Calendar validity, exchange holiday selection, source authenticity and provider receipt truth remain caller responsibilities as approved. Runtime wiring and new data collection are absent; future adapters require separate validation.

## Final snapshot SHA-256

| File | SHA-256 |
|---|---|
| internal/strategymarket/session_features.go | f965280c4f863a7971ef9479d9fa8f5310c42315d9e97b43100abc7b9ed6b495 |
| internal/strategymarket/session_features_test.go | 2a85d7a544d5465783056fec00e208a582b7a8503c96904ff1471068dfc520d9 |
| internal/strategymarket/SESSION_FEATURES.md | 72439c32dd66c24d057505803505e3ba4deb460057b127c7e91ee62bb637cb47 |

Pre-Landing Review: No unresolved issues found in this snapshot.

## Separate follow-up: existing a148 diagnostic boundary

At Manager request, assessed related existing a148 helpers without expanding a149 scope. **[P2, confidence 10/10]** `internal/strategyengine/rr_diagnostic.go` passes supplied entry/stop/quantity/FX/PnL/cost strings to positive/decimal/nonnegative; existing `lane.go:306` excludes only `eE/` and imposes no length limit. It therefore shares the p/P expansion issue above, with additionally unbounded literal length. An invalid record can incur large parsing allocation before returning invalid; unverified-provenance labels do not constrain resource use. There is currently no production caller of these diagnostic helpers, so this is a latent caller-input/resource-exhaustion issue, not evidence of deployed exploitation or live order risk. No payload executed. Recommend a separate scoped a148 diagnostic-only correction with strict bounded decimal validation and negative tests; do not change the shared runtime parser under a149. This inherited follow-up does not change the a149 verdict or claim the earlier a148 review covered this newly identified issue.
