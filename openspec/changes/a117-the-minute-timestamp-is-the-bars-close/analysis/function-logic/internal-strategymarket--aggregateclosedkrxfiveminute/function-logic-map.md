# Function Logic Map: `aggregateClosedKRXFiveMinute`

- Source: `internal/strategymarket/bars.go`
- Source SHA-256: `aa6c93cd24d872739e99ed9576fca8e80b3256c1b772f23737bc39fe39faab06` (current worktree, post-edit; verified with `sha256sum` 2026-08-18, equal to `source_sha256` in the regenerated `ast.json`). The pre-edit revision this analysis was written against was `a6f17d0cc8ede1b51d2e96344d124675d9022ed730142c50b9d06a6cc5cd9269`; branch identities are stable across the edit and only line numbers moved.
- Signature: `aggregateClosedKRXFiveMinute(market, symbol, source string, adjusted bool, raw []RawMinuteCandle, now time.Time) (VerifiedBar, error)` (`ast.json`: `params=6, results=2`)
- Source range: `131:1`-`228:2` (pre-edit `131:1`-`220:2`)
- AST counts: branches 18, returns 13, calls 37, defers 0, go statements 0 - unchanged by the edit (`ast.json` regenerated 2026-08-18 by `go run ./tools/logic-map` against the post-edit source).
- Risk scan: `risk-pattern-report.md`.
- **This function IS edited by a117.** Branches B7, B12 and the two time assignments inside the final return are the edit surface; every other branch must keep its exact refusal kind.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw` | exactly 5 `RawMinuteCandle` | `SealOfficialClosedKRXFiveMinuteFor`, which has already pinned market/symbol/interval/adjusted/source | B1 ⇒ `RefusalIncompleteBucket` |
| `candle.Timestamp` | RFC3339 with a numeric offset | the official `/api/v1/candles` response, passed through `official.RawMinuteCandles` without conversion | B5/B6 ⇒ `RefusalNaiveTimestamp` |
| `now` | non-zero, injected | caller | B2 ⇒ `RefusalOpenBucket` |
| KRX regular session | `09:00`-`15:30` KST | the a047 spec requirement (`openspec/specs/strategy-engine/spec.md:19`, *KRX open은 `09:00 KST`로 봉인*) | B7 ⇒ `RefusalOutsideRegularSession` |
| bucket alignment | five-minute grid anchored at `09:00` | same requirement | B12 ⇒ `RefusalIncompleteBucket` |

**The invariant this function currently gets wrong.** Every branch that reads a timestamp treats `candle.Timestamp` as the minute's **opening** instant. The broker's value is the minute's **closing** instant - measured live in both markets on 2026-08-18 and recorded as a112 decisions 30 and 31 (`openspec/changes/a112-run-four-strategy-families-independently/review.md`), and already acted on by `internal/officialbars/producer.go`, which converts `open_at = timestamp - 60s`. A bar labelled `t` therefore covers `[t-60s, t)`.

Three consequences follow, and all three are visible in the branch enumeration below rather than inferred:

1. **`openAt` is one minute late.** The final return sets `openAt: minutes[0].local` (line 210). The bucket actually opens at `minutes[0].local - 1m`.
2. **`closedAt` is one minute late.** Line 189 computes `minutes[0].local.Add(5 * time.Minute)`; the bucket actually closes at `minutes[4].local`, which is `minutes[0].local + 4m`.
3. **The session window admits and rejects the wrong minutes.** B7 requires `540 <= minuteOfDay < 930`. Under close labels the first regular minute is labelled `09:01` (it covers `09:00`-`09:01`) and the last is labelled `15:30` (it covers `15:29`-`15:30`). So the current code **admits a bar labelled `09:00`**, which covers the pre-open minute `08:59`-`09:00`, and **rejects the bar labelled `15:30`**, which is the session's last regular minute. B12 compounds it: the genuine first bucket starts with the bar labelled `09:01`, and `(541-540) % 5 == 1`, so the real first five-minute bucket of every KRX session is refused as unaligned today.

## Branches and early returns

Exact AST return nodes (post-edit): `133, 136, 140, 151, 155, 164, 170, 177, 183, 189, 193, 199, 213`; `183` is the `sort.Slice` comparator's own return, not an exit. Pre-edit the same thirteen nodes sat at `133, 136, 140, 151, 155, 160, 166, 173, 179, 182, 186, 191, 205`.

| Branch | Condition | Mutation/side effect | Return/error | Edited by a117 |
|---|---|---|---|---|
| B1 (`if` 132) | `len(raw) != 5` | none | `RefusalIncompleteBucket` | no |
| B2 (`if` 135) | `now.IsZero()` | none | `RefusalOpenBucket` | no |
| B3 (`if` 139) | `time.LoadLocation("Asia/Seoul")` failed | none | the raw `error` (not an `IntegrityError`) | no |
| B4 (`range` 149) | per input candle | builds `minutes` | - | no |
| B5 (`if` 150) | `offsetTimestamp` regexp rejects the literal | none | `RefusalNaiveTimestamp` | no |
| B6 (`if` 154) | `time.Parse` rejects it | none | `RefusalNaiveTimestamp` | no |
| B7 (`if` 163, pre-edit 159) | non-zero seconds/nanoseconds **or** `minuteOfDay < 9*60` **or** `>= 15*60+30` | none | `RefusalOutsideRegularSession` | **yes - window shifts to `9*60+1 .. 15*60+30` inclusive** |
| B8 (`if` 166, pre-edit 162) | first candle sets `currency` | assigns `currency` | - | no |
| B9 (`if` 169, pre-edit 165) | `Currency != "KRW"` or differs from the first | none | `RefusalCurrency` | no |
| B10 (`range` 174, pre-edit 170) | five decimal fields per candle | fills `values` | - | no |
| B11 (`if` 176, pre-edit 172) | `exactDecimal` rejected it or it is negative | none | `RefusalInvalidDecimal` | no |
| B12 (`if` 188, pre-edit 181) | `(startMinute - 9*60) % 5 != 0` | none | `RefusalIncompleteBucket` | **yes - alignment is computed from the derived open, not from the label** |
| B13 (`for` 191, pre-edit 184) | i = 1..4 | none | - | no |
| B14 (`if` 192, pre-edit 185) | `minutes[i] != minutes[0] + i minutes` | none | `RefusalMinuteGap` | no |
| B15 (`if` 198, pre-edit 190) | `now.Before(closedAt)` | none | `RefusalOpenBucket` | no (the predicate is unchanged; `closedAt`'s value changes) |
| B16 (`for` 204, pre-edit 196) | i = 0..4 | folds high/low/volume | - | no |
| B17 (`if` 205, pre-edit 197) | a higher high | `high.Set` | - | no |
| B18 (`if` 208, pre-edit 200) | a lower low | `low.Set` | - | no |
| final (213, pre-edit 205) | - | none | `VerifiedBar{openAt, closedAt, ...}` | **yes - `openAt` and `closedAt` values only** |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.LoadLocation` (138) | KST wall-clock window arithmetic | error propagates raw at B3 | `ast.json` call |
| `offsetTimestamp.MatchString` (150), `time.Parse` (153), `parsed.In` (157) | reject naive stamps, then convert to KST | refusals at B5/B6 | `ast.json` calls |
| `exactDecimal` (171), `decimalString` (212-216) | exact `big.Rat` decimal preservation, no float | refusal at B11 | `ast.json` calls |
| `sort.Slice` (179) | order the five minutes by instant | none | `ast.json` call |
| `SealOfficialClosedKRXFiveMinuteFor` (caller, `bars.go:115`) | pins identity/interval/adjusted/source before this function runs | typed refusals | CodeGraph: 3 callers, of which `AdaptAndSealClosedKRXFiveMinute` (`strategycandle/adapter.go:32`) and `SealOfficialClosedKRXFiveMinute` (`bars.go:111`) are the non-test ones |

**Production reach (CodeGraph, live index 34,752 nodes).** `AdaptAndSealClosedKRXFiveMinute` has exactly one caller and it is a test (`strategycandle/adapter_test.go:42`); `SealOfficialClosedKRXFiveMinute` has two, both tests (`strategyengine/lane_test.go:76`, `strategymarket/bars_test.go:52`). The only consumer of the resulting `VerifiedBar` is `strategyengine.LaneInput.Bar` (`contracts.go:143`), and no file outside `internal/strategyengine` constructs a `strategyengine.LaneInput`. The defect is therefore **latent**: today it can only mis-label bars inside tests. That is why a117 is a correction change and not an incident.

## State mutations and fallbacks

- No package state, no I/O, no clock read: `now` is injected. The function is pure given its inputs.
- There is no fallback path. Every failure is a typed refusal and no `VerifiedBar` is minted - the fail-closed shape is correct and a117 does not touch it.

## Safety conclusion

- Safe edit boundary: the edit changes three values (the session window bounds, the alignment anchor, and the two instants in the returned bar) and no refusal kind, no decimal arithmetic and no call. Every other branch keeps its current behaviour, which the existing table test pins.
- High-risk impact: **no** by the project's list - this function issues no order, stop, sizing, Guardian, ledger, reconciliation, auth or fill decision, and it has zero production callers today. It is nonetheless a strategy-input authority, so the change moves in the conservative direction only where it can: the pre-open minute that is admitted today becomes a refusal, and the session's true last minute becomes admissible. The correction cannot widen the session beyond `09:00`-`15:30`.
- Untested branch before a117: B3 (`LoadLocation` failure) and B9's second disjunct are unreachable in the current suite; every other branch is pinned by `TestAggregateClosedKRXFiveMinuteFailsClosed`'s eleven cases and `TestAggregateClosedKRXFiveMinutePreservesExactDecimals`. a117 adds REDs for both edited branches in both directions.
