# a143 independent adversarial implementation review

Reviewer: `/root/a143_adversarial_reviewer` (independent from Terra implementer).
Date: 2026-10-16 KST. Baseline: `71e9bae1`; frozen implementation: `926d67a2`.
Verdict: **CHANGES REQUESTED — P2 regression coverage**. No production correctness defect found in reviewed diff. No live API call, operating change, source edit, or test edit performed by reviewer.

## Finding

**P2: multi-candle test does not prove the frozen preservation contract.** `TestAdaptCandlesOrdersNewestFirstWireChronologically` uses two rows containing only Close and Timestamp. It checks Close order and relative timestamp order. Frozen spec requires three descending candles and unchanged timestamp/price/volume values. Incorrect per-row timestamp shifts, value association mistakes outside Close, or a special-case two-row reversal could escape this regression. Existing full-value singleton coverage does not exercise reordered rows.

Requested correction: use three candles with distinct full OHLCV and exact timestamp values; assert every output row equals its corresponding expected chronological row, retaining the existing empty/singleton tests. This is a test change only; current full-struct swap is correct by inspection.

## Verified scope

Only two production files differ from baseline: general `adaptCandles` adds a bounded in-place reversal of the newly adapted slice; KR strict regex changes from six digits to six ASCII uppercase alphanumeric characters. No strict reader timestamp, raw ordering, cursor, price validation, normalization or US lexical implementation changed. `officialbars`, strategy/order/Guardian/journal/token paths, workflows and operating configuration are absent from the diff.

Production SHA-256 at review:

- candle_reads.go: `d5be79e9caef683c3f6c8f075aad50497c64ebff39040419579494bfd3fbca1a`
- strict_minute_candles.go: `d511149cbf1bdd7f83646fd7863507a74840a38c69fa6c6a68705b72b1549163`

Empty/singleton reversal boundaries are safe; swapping complete domain.Candle values preserves timestamp and OHLCV. The regex is anchored, ASCII-only and case-sensitive; whitespace/punctuation/non-ASCII remain refused. New strict test verifies exact `0101N0` request query and returned Page.Symbol; it does not falsely assert broker-response identity.

## Independent offline snapshot checks

Recomputed with installed jq and sha256sum, not copied from metadata:

| Snapshot | Version | Count | Canonical full SHA-256 |
| --- | --- | --- | --- |
| REST | 1.2.24 | 38 paths | 91f9e02f79df606f9a02db23f391ff993f8e634417a57fd2f5ef9d9d188308bd |
| AsyncAPI | 1.2.2 | 4 channels | 176595a3abbba3c1b84576e3c1b6806b75ec6d507a2efee26a24021a55531a63 |

REST paths hash: `5078c1acd2bb5ea1d7560ce37c8cbcaec3a5795363835955b48348cd97e934cb`. Full hashes use exact `jq -S .` output with trailing newline; paths hash uses `jq -S '.paths | keys'`. All match stored metadata. REST metadata preserves existing fields and monitor format. AsyncAPI metadata explicitly says snapshot-only/no automatic monitoring. Public source URLs are recorded. Retrieval date/HTTP status provenance is Terra's receipt responsibility; reviewer did not independently refetch those URLs.

## Test evidence and limits

Read raw `/tmp/a143-red.log`: both new behavioral tests fail on prior code for the expected ordering and lexical rejection reasons. Read `/tmp/a143-green-official.log` and `/tmp/a143-green-officialbars.log`: package test success receipts exist. Reviewed existing strict refusal/query tests, singleton/empty chart assertions, and officialbars invariant test inventory including cursor progression, successor tracking and close-label conversion. These logs are implementer-produced, not independent reviewer test runs. No duplicated full suite executed.

Analysis documents were being refreshed concurrently; this review does not certify their final freshness, required mutations, full gate, additional gstack review or Manager acceptance. After test correction, re-review frozen diff and record the focused successful test receipt before changing verdict to PASS.

## Final re-review — 60f1d041

**PASS.** P2 resolved. Compared `926d67a2..60f1d041`: only candle_reads_test.go changes. Regression now supplies three uniquely valued full OHLCV candles and compares the complete chronological domain.Candle slice against independent explicit expected values and exact parsed timestamps. Existing singleton, empty and bad-timestamp tests remain present. Production/snapshot conclusions above remain tied to unchanged `926d67a2` contents.

Reviewer independently executed offline focused check:

```text
go test ./internal/official -run 'TestAdaptCandles' -count=1
ok github.com/JungHoonGhae/tossinvest-cli/internal/official 0.002s
exit 0
```

No outstanding actionable code/test finding. Snapshot hashes/counts were independently verified offline; live source retrieval status/date provenance was not independently re-fetched and remains subject to Terra's explicit receipt. Concurrent analysis/receipt finalization, additional gstack review, required full gate and Manager completion verification remain outside this PASS.

## Final target update — e6718bc17e91086b189e05c2319fbd95a7999fe9

**PASS retained.** Inspected test-only commit `e6718bc1`: one table row adds pure seven-character KR symbol `0101N00` to the existing pre-request refusal test; associated AST is refreshed. No further test execution duplicated. Compared production sources and snapshot bodies from `60f1d041` to this target: no changes; both production SHA-256 values still match the values above.

Also inspected `2d8b5f86`: only AsyncAPI metadata source URL changes to `https://openapi.tossinvest.com/openapi-docs/latest/asyncapi.json`. Snapshot bytes, version, count and canonical hash remain unchanged. Read Terra's verification receipt reporting HTTP 200 at `2026-10-15T16:44:26Z` and byte equality to stored AsyncAPI. This is reviewed implementer retrieval evidence, not a new reviewer network fetch. Final strict/adapt/officialbars focused success is recorded in that receipt; earlier independent reviewer adapter execution remains the independent test evidence.

No new actionable finding. Verdict covers final code/tests and metadata correction through `e6718bc1`; it does not substitute for additional gstack review, full gates or Manager completion verification.
