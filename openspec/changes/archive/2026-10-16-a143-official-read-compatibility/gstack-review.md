# a143 additional gstack review

2026-10-16. Reviewer `/root/gstack_compatibility_reviewer`, independent Codex context after dedicated adversarial review. Applied gstack review skill/checklist and combined specialist/red-team perspectives. Baseline `71e9bae1`, production/snapshots `926d67a2`, final tests `60f1d041`, analysis `4207c865`, AsyncAPI URL correction `2d8b5f86`.

## Verdict

**PASS REVIEW for corrected dafa324b implementation and final evidence.** Shared-caller regression and snapshot-test mismatch are corrected and re-reviewed below. Earlier withdrawn PASS and failed aggregate gate remain in this record. Full gate/completion is not asserted.

## P2 finding — final analysis does not match AST/test evidence

Confidence 10/10. In `analysis/function-logic/`:

- `internal-official--adaptcandles/function-logic-map.md` omits current AST B4 reversal and says planned edit. BTM B2 is timestamp-nonempty, not empty input; B3 is parse-success, not invalid-input branch. Singleton invariant cites the three-candle test rather than `TestAdaptCandlesUnit`.
- `internal-official--client.strictminutecandles/branch-test-map.md` assigns numeric acceptance to B4 (actual count refusal), invalid KR to B5 (before nonempty), US policy to B6 (before parse error) and raw-order/cursor to B7 (before query). New alphanumeric test has only one candle and cannot prove newest-first/cursor behavior. Use actual AST IDs and appropriate existing raw-page tests, separating static unchanged paths from executed evidence.
- `internal-official--strictminutechecksymbol/branch-test-map.md` links B2 predicate-failure only to numeric success. Include invalid-form rejection and distinguish true/false outcomes.
- Pending RED/GREEN placeholders do not reflect final receipts. The invalid-argument test BTM puts focused GREEN in RED column. Do not imply instrumented branch coverage from package/test success.
- Unchanged neighboring test ASTs have historical source hashes (`TestAdaptCandlesEmpty`: 78fb..., KoreanMarket: b4f729...) versus current files; identify their baseline nature in FLM instead of presenting as final-source artifacts.

Sent consolidated corrections to Manager; documentation-only repair, no product/test change requested.

## Source and test review

Only general adaptCandles reverses the newly allocated complete domain.Candle slice. Empty/singleton termination is safe; full-struct swaps retain associated OHLCV/time values. StrictMinuteCandles and officialbars bypass this adapter. Domestic regex alone expands to six anchored ASCII uppercase alphanumeric characters; US predicate, case sensitivity, request propagation, strict raw ordering/cursor/timestamp/price validation are unchanged. Three-candle regression compares full explicit values/timestamps; invalid KR character cases and exact query/Page.Symbol are present. The latter is request identity preservation, not broker-response identity attestation.

Read full scoped Go diff, caller separation, current AST branch inventories, tests and dedicated review receipts. No SQL/persistent data, shell execution, concurrency state or LLM execution boundary change. No workflow, strategy, order, Guardian, journal, token or operating setting source changed. No tests/live calls executed by this reviewer.

Fresh independent `/root/gstack_compatibility_reviewer/a143_specialist` applied testing/security/API-contract/maintainability/performance/red-team checklists in one context and returned NO FINDINGS. It confirmed adapter/strict/officialbars separation. This is one combined specialist, not six independent models; actual reviewers are Codex agent contexts. No Claude or separate Codex CLI process, remote fetch/PR/Greptile/version-queue, telemetry, setup, external sync or memory writes performed. Full suites were not duplicated.

## Snapshot verification

Independently recomputed local jq canonical hashes, including jq's trailing newline, and counts:

| Document | Version/count | SHA-256 |
|---|---|---|
| REST full | 1.2.24 / 38 paths | 91f9e02f79df606f9a02db23f391ff993f8e634417a57fd2f5ef9d9d188308bd |
| REST paths | jq -S '.paths \| keys' | 5078c1acd2bb5ea1d7560ce37c8cbcaec3a5795363835955b48348cd97e934cb |
| AsyncAPI full | 1.2.2 / 4 channels | 176595a3abbba3c1b84576e3c1b6806b75ec6d507a2efee26a24021a55531a63 |

All equal metadata. REST monitor schema retained. AsyncAPI spec_url matches the receipt's actual `https://openapi.tossinvest.com/openapi-docs/latest/asyncapi.json` retrieval URL; receipt gives HTTP 200 at 2026-10-15T16:44:26Z. Retrieval/date/status evidence is Terra-produced, not independently refetched by this reviewer. AsyncAPI explicitly states snapshot-only; no automated monitoring or WebSocket capability claimed.

## Reviewed SHA-256

```text
internal/official/candle_reads.go d5be79e9caef683c3f6c8f075aad50497c64ebff39040419579494bfd3fbca1a
internal/official/candle_reads_test.go 87b7975c7caa06c02311b2d9fdbde19422cf72c55f5fcfaf07582ca7e8421a9c
internal/official/strict_minute_candles.go d511149cbf1bdd7f83646fd7863507a74840a38c69fa6c6a68705b72b1549163
internal/official/strict_minute_candles_test.go d5312e14348516437f5852f5e5b0f94096bb71f843d2eeb78173b4ffb7062f77
```

## Final closure re-review after e6718bc1

P2 CLOSED. Re-read Manager's final FLM/BTM: adaptCandles enumerates actual B1–B4 and ties singleton to TestAdaptCandlesUnit; StrictMinuteCandles B1–B11 match the current AST and explicitly leave final-attempt error coverage unisolated; strictMinuteCheckSymbol selection/mismatch conditions match actual B1/B2. Historical neighboring-test ASTs are identified as baseline evidence. RED, mutation and named-test PASS claims are distinguished without instrumented branch-coverage claims. New one-candle KR fixture is limited to exact request/Page.Symbol; the existing two-candle canonical-query fixture supplies raw newest-first timestamp/cursor assertions.

Inspected e6718bc1 diff: one seven-character domestic-symbol refusal case added, with no product change. Read three final focused exit receipts (adapt, strict, officialbars: all 0) and Manager's final analysis checker output (evidence complete/diff-proven exempt). These are execution receipts from other roles, not tests run by this gstack reviewer. Final strict test hash is updated below; remaining three Go hashes and snapshot contents are unchanged. Source scope remains two production files and two test files plus public snapshots/metadata and OpenSpec evidence. No workflow/operating settings change.

No unresolved additional gstack finding remains. Full aggregate gate, archive and Manager completion verification remain pending/separate; review PASS does not imply their success.

Final strict test SHA-256 at e6718bc1:

```text
internal/official/strict_minute_candles_test.go a8b64d26d5ed9ac7c0d92cbd8f3afd9297d9e8736c92c2d45ea95b08321e49c8
```

## Full-gate regression — prior PASS withdrawn

Manager reports full gate failures after the above focused review: strict quote's existing numeric-only KR symbol contract changed through its call to the shared strictMinuteCheckSymbol helper (`internal/official/strict_quote_reads.go:201`), and the a087 public-contract text assertion failed against the refreshed snapshot. The previous review's caller analysis was incomplete: it established general/strict candle separation but did not trace all callers of the shared lexical helper. Its statements that the lexical expansion affected only the minute reader must not be used as verified facts. Focused strict-minute/officialbars PASS receipts did not prove unchanged strict-quote behavior.

Current blocking issue: isolate the minute-only lexical expansion while preserving strict quote's established numeric-only KR contract, under Manager's amended specification, then run the affected quote/minute fixtures and full gate. The a087 assertion requires separate examination of whether only public description wording changed; no behavior-policy weakening is authorized by this review. Terra owns fixes/tests and the dedicated reviewer must inspect them before additional gstack re-review. Reviewer has changed only this review record, not product/tests. Earlier source hashes identify the reviewed-but-regressing snapshot, not approved final implementation.

## Corrected implementation final re-review — dafa324b / a447ccf0

**Regression CLOSED; PASS REVIEW restored for this corrected snapshot only.** Read the correction diff and searched all repository Go references to strictMinuteCheckSymbol, strictMinuteCheckCandleSymbol, strictMinuteKRSymbol and strictMinuteCandleKRSymbol. StrictMinuteCandles alone calls the new minute wrapper. The wrapper accepts KR six-character uppercase ASCII alphanumeric inputs and otherwise delegates to the restored shared numeric-KR/unchanged-US helper. strictQuoteIdentity still calls that shared helper directly; strict quote production code is unchanged. The same 0101N0 is accepted by minute fixtures but refused with exact StrictQuoteReasonSymbol by both quote readers, with zero broker requests. Numeric quote success is proven by the actual KR 005930 fixture, not by a minute fixture that bypasses the shared helper.

The a087 diff changes one public-description expectation string only. Existing request-body assertions and executable order policy are unchanged; no new order permission follows from the documentation string. Read final minute, full strict-quote, numeric-quote, complete a087 package and analysis exit files: all 0. These are Terra/reviewer-produced receipts; this gstack reviewer did not repeat tests or access services.

Read current wrapper AST/FLM/BTM and corrected shared-helper mappings. Wrapper has one B1 with true/false paths, not invented separate fallback branch IDs. Valid US fallback cites the actual canonical query/page test. Historical artifacts remain historical; earlier missed caller and failed gate are not relabeled. General candle reversal, raw strict/time/cursor handling, officialbars, US policy and snapshot bytes remain unchanged by this fix.

Final Go scope: two production files, three official-package test files and one a087 test file. No quote production, strategy/order execution, Guardian, journal, token, workflow or operating configuration changes. Final changed hashes versus the earlier record:

```text
internal/official/strict_minute_candles.go 8f6c64f5f72feff0579de6f42045d8861e710209da2f8ef5376fe827cd047185
internal/official/strict_quote_reads_test.go 5ac17738f04cd6b633b6e8d1bd12a9d6147ad81b9f6a9bd100a56491508e51f2
tools/a087-market-sell-probe/contract_test.go 5864c940884a26000327d5c0a442ab133274c74803ab17d95aa0d9affdfa521f
```

The additional independent specialist also explicitly withdrew its earlier incorrect NO FINDINGS for the missed quote caller, then re-reviewed dafa324b with repository-wide references and returned NO FINDINGS for the corrected scope. This was static review only, with no tests/network/writes.

Current candle source/test and strict-minute test hashes remain as recorded in the previous final-target section. No actionable new finding in the corrected scope. Required full aggregate gate must run successfully against the corrected final state; previous failed gate and focused successes cannot substitute. This review does not certify live API behavior, deployment or completed integration.
