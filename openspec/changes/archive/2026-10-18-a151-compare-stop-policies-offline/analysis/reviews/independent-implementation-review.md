# Independent a151 implementation review

Verdict: IMPLEMENTATION REVIEW PASS. Three original P2 findings below are resolved in source 3d7849652be408103211883888b9b6ef7a5828693f9f51d51b4408c07bbecc54. Initial FAIL applied to source 29288f99aeee437fd21d5ee569397350fec2b33db7c3bfc6b97cc16097a57ccb; history retained. Aggregate gate acceptance is separate. Reviewer performed source/test/evidence inspection only; no tests or live operations executed.

## P2: Post-entry invalid exit erases pre-entry eligibility and sibling outcomes

In evaluate arm replay, invalid_exit_price raises CandidateError. compare replaces the complete candidate with an invalid status/reason, dropping validated entry/geometry and all arm outcomes. Market eligible counts then exclude it. This conflates pre-entry eligibility with post-entry completion, violating separate cohort semantics. Existing test_invalid_exit_and_missing_path asserts the whole-row invalid result and therefore cannot detect this defect.

Counterexample request: valid US fixture with exit_slippage_ticks=1000. Valid sizing precedes a nonpositive exit fill. Preserve eligible=1 and complete=0; represent affected arm outcome as invalid and preserve other closed/censored arm results. Add assertions for denominator and retained outcomes, not merely an exception/status.

## P2: Future redundant five-minute bars convert path censoring into pre-entry exclusion

The all-five-bars aggregation loop precedes signal-prefix and entry handling. A supplied future five-minute bar missing one constituent minute raises missing_1m_aggregate and discards the candidate. Without that redundant five-minute bar, the identical missing post-entry minute censors the arm path. Cohort membership thus depends on optional outcome aggregate data. Restrict required completeness to signal prerequisites; future missing constituents must preserve post-entry censoring. Present contradictory OHLC aggregates may still be rejected. Verify equivalent minute gaps with and without future five-minute evidence have identical eligible/complete counts.

## P2: Whole-session tick validity does not cover warmup

README declares a single grid for the whole session and code uses it to validate all warmup bars. Yet tick_scope_or_validity only requires valid_from <= signal and valid_until >= session close. valid_from=signal > session open is accepted while earlier prices are interpreted using an out-of-period grid. Enforce valid_from <= session open for this declared scope; retain valid_until coverage. Request valid_from=signal regression with otherwise unchanged fixture.

## Reviewed correct boundaries and evidence limits

Independent arithmetic agrees with source geometry: KR entry10000/tick10/low9900/ATR80/budget7000 yields stops9930/9890/9880, quantities100/63/58, target10210; equivalent native USD inputs yield same dimensionless R. Wilder seed/update, fee ceiling, entry-bar gap precedence, stop-first ambiguity with independent target-first sensitivity, and complete-ID paired mean formulas match approved definitions. UTC-aware session comparisons validate US offsets against IANA; signal receipts and snapshot dates are checked. Output creation walks directories using O_NOFOLLOW and exclusive final creation, protecting existing targets. No network/live imports appear in reviewed source.

Read 20-test PASS receipt and CLI evidence; synthetic fixture contains KR and US, 77 minute bars and 15 five-minute bars per candidate, with deliberate final both-touch bar. These establish synthetic coverage only. Full aggregate gate has not run at this review. Existing focused success does not override the findings above. No real-data performance or operational approval claimed.

## Original reviewed SHA-256

- `tools/a151-compare-stop-policies-offline/analysis.py`: `3d7849652be408103211883888b9b6ef7a5828693f9f51d51b4408c07bbecc54`
- `tools/a151-compare-stop-policies-offline/test_analysis.py`: `3d96717180cacaf8a998714d01e6f4749b898af897b979e2fee6da8295bade89`
- `tools/a151-compare-stop-policies-offline/README.md`: `e5b140f79f6d5f1846c7899a059f013838e8ae56e0ab909e81ef1cc04addee7e`
- `tools/a151-compare-stop-policies-offline/testdata/kr-us-synthetic.json`: `3f50b20bd5b108f1c398cbcd606184b2af2de99da5111f1c92df9338ecfe3f50`
- `openspec/changes/a151-compare-stop-policies-offline/design.md`: `7090bcce06221e889e6455359394911acc4232a2d26c452cfd7ceac8c2e08e38`
- `openspec/changes/a151-compare-stop-policies-offline/specs/strategy-optimization/spec.md`: `2bd0ee278f988450e6a319315917942a9551d4716600b7bc5291b549e3fd4662`
- `openspec/changes/a151-compare-stop-policies-offline/analysis/verification/implementation-evidence.md`: `4ddc0b75c151a9559f1d0687775ce168af3d936fdbf0fac8c2e6b59429b3c0bc`
- `openspec/changes/a151-compare-stop-policies-offline/analysis/verification/focused-tests.log`: `e9010de26ea5b87cc8ad6b28a52fd6f977bf9090a3bb357f27569d20a42d5703`

## Resolution rereview

All three fixes match the Manager clarification. Invalid exits now become local per-arm/per-order outcomes, retaining eligible geometry, other arms and independent sensitivity. Statistics count invalid arms and require all base arms closed for complete comparisons. Strengthened test checks invalid ATR base alongside closed ATR sensitivity and closed fixed base, with complete=0.

Future five-minute aggregates with missing constituents are explicitly unverified and no longer cause pre-entry exclusion. The minute path still stops at first missing bar. Present complete constituent groups continue to be checked; a new negative test rejects contradictory volume. Whole-session tick validation now requires valid_from<=open and valid_until>=close.

Inspected behavioral RED receipt /tmp/tossos-a151-reviews-red.log: three failures demonstrate the original issues. Inspected /tmp/tossos-a151-reviews-green2.log: 23 tests PASS. Reviewed changed source/test/README and Manager design clarification. No remaining actionable implementation finding. Reviewer did not execute tests. Real-market validity and profitability remain unverified/inconclusive; full aggregate gate still pending.

## Final reviewed SHA-256

- `tools/a151-compare-stop-policies-offline/analysis.py`: `3d7849652be408103211883888b9b6ef7a5828693f9f51d51b4408c07bbecc54`
- `tools/a151-compare-stop-policies-offline/test_analysis.py`: `3d96717180cacaf8a998714d01e6f4749b898af897b979e2fee6da8295bade89`
- `tools/a151-compare-stop-policies-offline/README.md`: `e5b140f79f6d5f1846c7899a059f013838e8ae56e0ab909e81ef1cc04addee7e`
- `tools/a151-compare-stop-policies-offline/testdata/kr-us-synthetic.json`: `3f50b20bd5b108f1c398cbcd606184b2af2de99da5111f1c92df9338ecfe3f50`
- `openspec/changes/a151-compare-stop-policies-offline/design.md`: `7090bcce06221e889e6455359394911acc4232a2d26c452cfd7ceac8c2e08e38`
- `openspec/changes/a151-compare-stop-policies-offline/specs/strategy-optimization/spec.md`: `2bd0ee278f988450e6a319315917942a9551d4716600b7bc5291b549e3fd4662`
