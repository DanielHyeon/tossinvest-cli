# a151 gstack implementation review

Verdict: scoped implementation review PASS after corrective rereview. No unresolved actionable finding remains. Full aggregate gate is pending and is not covered by this PASS. Codex-host gstack critical/informational passes with security and QA inspection; no Claude/cross-model run claimed. Reviewer read source, tests, README, synthetic manifest/output, approved OpenSpec contracts and implementer evidence. Reviewer executed no code under review, tests, CLI or gates; only this report edited. Full aggregate gate has not been established.

## Initial findings — resolved (history preserved)

1. **P2: post-entry missing data becomes pre-entry exclusion** (`analysis.py:151–154`). Validation loops over every supplied five-minute bar before entry and requires all one-minute constituents. A valid signal/entry followed by a missing minute and a supplied later five-minute aggregate returns `missing_1m_aggregate`; no eligible arm or censored outcome remains. Contract requires censoring at the missing path. Keep strict prefix validation for signal inputs, but do not let optional future five-minute availability retroactively remove the eligible trade. Negative fixture: signal at minute75, entry open76, missing minute78, supplied five-minute close80; preserve earlier arm outcomes and censor still-open arms.
2. **P2: invalid exit collapses pre/post cohorts** (`analysis.py:191,213,217`). An invalid computed fill raises out of the whole candidate after entry, discarding other arm outcomes and removing an originally eligible observation from `eligible`. Keep per-arm invalid outcome/reason and pre-entry eligibility, exclude the candidate only from complete paired comparison. Negative fixture should allow one arm to close normally while another later stop/slippage fill is invalid/out of the declared grid. Assert eligible=1, complete=0 and preservation of the first arm. Existing `test_invalid_exit_and_missing_path` currently expects collapsed candidate invalidity and must be aligned with the cohort contract.

Findings sent to Manager/Terra and reproduced by implementer in the retained RED receipt. No reviewer-run reproduction is claimed.

## Corrective rereview

Both P2 findings are resolved. Future five-minute bars with missing constituents now produce explicit unverified metadata; strict required signal-prefix checks remain. Execution follows the contiguous one-minute path, preserves prior closed arms, and censors still-open arms at the gap. Fully available contradictory aggregates remain rejected. Post-entry invalid fill is now a per-arm/per-sensitivity outcome; eligibility and siblings remain, invalid base count is reported, and complete comparison excludes it. The test deliberately yields fixed/base closed, ATR/base invalid, ATR/sensitivity closed, confirming separation.

Also verified separate reviewer finding: tick valid_from must cover session open, not merely confirmation, and valid_until covers session close. The approved design and README explicitly reflect these contracts.

Read `/tmp/tossos-a151-reviews-red.log`: three actual behavioral failures. Read `/tmp/tossos-a151-reviews-green2.log`: 23 tests PASS. Copied review-findings receipts preserve this history; initial missing-module RED is not confused with behavioral RED. Read fresh CLI output `kr-us-synthetic-output-after-review.json`, verified its new invalid counts/unverified fields, and both markets retain one eligible/complete synthetic candidate. Output SHA-256 `cfb6bb5ba79bfd82a2cd26bc75280ad8034a8122a780cf88ad6a9377cbeb2af2`. Earlier 20-test/CLI receipts are explicitly previous-snapshot evidence. No tests or CLI executed by reviewer.

## Inspected strengths and limits

- Bounded decimal grammar precedes Fraction parsing; exact stop geometry, tick rounding, integer sizing, side fees and tax match declared formulas. Zero caps do not mean unlimited. Exact output rationals avoid floating-point PnL.
- Signal prefix is contiguous and requires close/receipt availability; ATR uses anchor plus fourteen true ranges and Wilder continuation. Next scheduled open is strictly after confirmation. Future HLC does not enter stop/sizing arithmetic.
- Session endpoints validate IANA offset/date, KR/US regular/early-close boundaries. USD/KRW budgets, tick scope/validity and explicit costs are enforced; no FX or live market authority implied.
- Open-gap precedence includes entry bar. Same-bar stop-first versus target-first is explicit and disclosed as arm-specific sensitivity, not joint paired bounds. Main comparison uses all-three closed IDs and null empty means; corrected branches preserve eligibility and outcomes before this filter.
- Exclusive output uses directory-descriptor traversal with O_NOFOLLOW and O_CREAT|O_EXCL. No existing file/input/symlink target overwrite path found. This is exclusive creation, not a guarantee that a newly created file is removed after a disk-write failure; no stronger durability claim made.
- No network, subprocess, live trading, credential or runtime imports in implementation. JSON duplicate keys/constants and unknown fields are rejected. Input bytes are hashed and left untouched. No new dependencies or production modifications needed.
- Read 20-test PASS receipt in `analysis/verification/focused-tests.log`; read CLI synthetic output evidence. Tests include independent arithmetic, DST, caps, gaps, input preservation and symlink refusal. These tests are explicitly outside aggregate Makefile discovery and must run separately. Synthetic output does not establish real-data quality or profitability. Full gate remains separate.

## Initial reviewed source

`analysis.py` SHA-256: `29288f99aeee437fd21d5ee569397350fec2b33db7c3bfc6b97cc16097a57ccb`.
Initial source above is superseded by the corrected manifest below.

## Final reviewed manifest

Source/tests/docs/fixture and reviewed receipts; excludes this report and other reviewers. Later changes require refresh.

Pre-gate metadata refresh: only tasks.md and implementation-evidence.md changed. Completed boxes describe review/preparation; final sync/check, aggregate gate, archive and commit/push retain separate actual-receipt requirements. Source, tests, README, fixture, design/spec and stored behavioral receipts/output hashes unchanged. Prior manifest `1c2af4d32d2fa3d4ef5d4f21f525cbf445b4c303f48ff271a4c8cb3005f2e6af` superseded.

Manifest SHA-256: `1efb8818abbb25bdfcdd0b9b484c94d42e38438272a19b1622ce792943913cc2`

```text
3d7849652be408103211883888b9b6ef7a5828693f9f51d51b4408c07bbecc54  tools/a151-compare-stop-policies-offline/analysis.py
3d96717180cacaf8a998714d01e6f4749b898af897b979e2fee6da8295bade89  tools/a151-compare-stop-policies-offline/test_analysis.py
e5b140f79f6d5f1846c7899a059f013838e8ae56e0ab909e81ef1cc04addee7e  tools/a151-compare-stop-policies-offline/README.md
3f50b20bd5b108f1c398cbcd606184b2af2de99da5111f1c92df9338ecfe3f50  tools/a151-compare-stop-policies-offline/testdata/kr-us-synthetic.json
380f9ca931c674484a064752e1dc6da72bea38cc442a2567c44975c796618468  openspec/changes/a151-compare-stop-policies-offline/proposal.md
7090bcce06221e889e6455359394911acc4232a2d26c452cfd7ceac8c2e08e38  openspec/changes/a151-compare-stop-policies-offline/design.md
2bd0ee278f988450e6a319315917942a9551d4716600b7bc5291b549e3fd4662  openspec/changes/a151-compare-stop-policies-offline/specs/strategy-optimization/spec.md
fab86f24efa99a6da8a4b3f9a888a020552ea0c93ba21c7c2c74be9bfc045683  openspec/changes/a151-compare-stop-policies-offline/tasks.md
30b28e1562362a51d37ed07d7dc9f38717c0a8e8bcf44ebc72547107373720c7  openspec/changes/a151-compare-stop-policies-offline/analysis/verification/implementation-evidence.md
f57c571d85ac32275fa6ce752d46571d6f9b23b7a9f09940da2bf404fd4d0051  openspec/changes/a151-compare-stop-policies-offline/analysis/verification/review-findings-red.log
76950cb07db694a8e857cef792fabbc4a39d63528676388b1869961172a951c6  openspec/changes/a151-compare-stop-policies-offline/analysis/verification/review-findings-green.log
cfb6bb5ba79bfd82a2cd26bc75280ad8034a8122a780cf88ad6a9377cbeb2af2  openspec/changes/a151-compare-stop-policies-offline/analysis/verification/kr-us-synthetic-output-after-review.json
```
