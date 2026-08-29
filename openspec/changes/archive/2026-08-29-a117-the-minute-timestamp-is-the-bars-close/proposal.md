## Why

The Toss Open API returns each 1-minute candle with a `timestamp`, and the broker's own document calls it the bar's opening instant. It is not. Live probes on 2026-08-18 measured the convention in both markets — Korea with `005930` at 09:06 KST and the US with `AAPL` — and both showed the same thing: the label is the instant the bar **closed**. A bar labelled `t` covers `[t − interval, t)`. That is a112 decisions 30 and 31, and `internal/officialbars` already converts (`open_at = timestamp − 60s`).

Two copies of the old, false reading survive in this repository and are not covered by a112's ownership boundary:

1. `internal/official/candle_reads.go:16` states `— bar open time` in the DTO's schema comment. Documentation only, but it is the first thing the next reader of that file believes.
2. `internal/strategymarket.aggregateClosedKRXFiveMinute` **acts** on it. Its AST enumeration (`analysis/function-logic/internal-strategymarket--aggregateclosedkrxfiveminute/ast.json`, 18 branches) shows three places: the regular-session window test (B7), the five-minute grid alignment test (B12), and the two instants written into the returned bar. Every KRX five-minute bar it produces is therefore shifted one minute late, the genuine first bucket of each session (`09:01`–`09:05` labels) is refused as unaligned, the pre-open minute `08:59`–`09:00` is admitted as regular, and the session's true last minute (`15:30` label) is refused.

The defect is latent, and this proposal says so plainly rather than claiming an incident: CodeGraph shows the aggregator's only callers are tests, and no file outside `internal/strategyengine` constructs the `LaneInput` that consumes the result. It is corrected now because the correction unit is the *value*, not the file — leaving an executable copy of a claim we have measured to be false is how the same error comes back.

A third copy lives in `docs/migration/openapi.latest.json`. That file is the broker's document, mirrored here for reference; a117 records it and does not edit it.

## What Changes

- Derive the bucket's opening instant from the first label (`label − 1 minute`) instead of using the label directly, and anchor the five-minute grid check on that derived instant.
- Move the regular-session admission window from "label in `[09:00, 15:30)`" to "label in `[09:01, 15:30]`", which is the same wall-clock session (`09:00`–`15:30` KST) read through close labels.
- Set the returned bar's `closedAt` to the derived open plus five minutes, which is exactly the last label.
- Correct the DTO schema comment in `internal/official/candle_reads.go` and say why the broker's document is wrong there.
- Shift the three affected test fixtures by one minute so each keeps meaning the same bucket; every existing refusal case and its typed reason is unchanged.

No refusal kind, decimal arithmetic, call, signature or export changes. No runtime is wired, no toggle moves, no container is replaced.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `strategy-engine`: the frozen KRX five-minute input requirement now states the label convention it depends on.

## Impact

- Corrected implementation: `internal/strategymarket/bars.go`
- Corrected documentation: `internal/official/candle_reads.go`
- Fixtures shifted by one minute (same buckets): `internal/strategymarket/bars_test.go`, `internal/strategycandle/adapter_test.go`, `internal/strategyengine/lane_test.go`
- Not edited: `docs/migration/openapi.latest.json` (broker-owned), `internal/officialbars/**` (already correct), any order, stop, sizing, Guardian, ledger, reconciliation, auth or fill path.
