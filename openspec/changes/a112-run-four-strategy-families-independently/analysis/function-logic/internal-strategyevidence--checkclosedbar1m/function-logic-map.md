# Function Logic Map: `checkClosedBar1m`

- Source: `internal/strategyevidence/breakout_bar.go`
- Source SHA-256: `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `checkClosedBar1m(bar ClosedBar1mPayload) error`
- Source range: `464:1`–`571:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 28, returns 27, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Called only from `decodeClosedBar1mObject` (breakout_bar.go:458) after the strict field reader (`reader.done()` — unknown/missing/mistyped fields already refused). Every input field is therefore present and of the declared JSON type; this function checks values and cross-field relations only.
- The reader runs on the canonical object (`decodeCanonicalObject` → `canonicalJSON`), so integers are already value-checked by `canonicalIntegerValue`; raw price strings are untouched (lossless).
- Pure function: no store, clock or network access. Refusals are `payloadFieldError(kind, field, detail)` naming the offending field; the first failing rule wins (order is load-bearing for the refusal detail that tests assert on).
- Scale table single source: `marketMoney` (KR→KRW/0, US→USD/4; amended decision 9).

## Branches and early returns

Exact AST return nodes: `467, 471, 474, 477, 480, 484, 487, 490, 493, 496, 499, 502, 506, 512, 516, 519, 527, 531, 534, 537, 551, 555, 559, 562, 565, 568, 570` (26 refusals + final `nil`).

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 466:2 | `schema` ≠ `official_closed_bar_1m:v1` → refuse | `TestClosedBarRejectsUnknownEnumValues` /schema |
| B2 | if | 470:2 | `market` not in `marketMoney` (KR/US) → refuse | `TestClosedBarRejectsUnknownEnumValues` /market |
| B3 | if | 473:2 | `currency` ≠ the market's currency → refuse | `TestClosedBarRejectsUnknownEnumValues` /currency |
| B4 | if | 476:2 | `price_scale` ≠ the market's scale → refuse | `TestClosedBarRejectsOverPreciseRawForTheDeclaredScale` (declared 0/2/3/5 for USD) |
| B5 | if | 479:2 | `symbol` not canonical stored text (upper-case, no `:`, no space/NUL, non-empty) → refuse | `TestClosedBarRefusesSymbolWithRecordSeparatorOrLowerCase`, `TestClosedBarRejectsMalformedIdentityStrings` |
| B6 | if | 483:2 | `session_id` not `<CAL>:<date>` for the market → refuse | `TestClosedBarRejectsMalformedIdentityStrings`, `TestClosedBarRejectsUnknownEnumValues` /session_id |
| B7 | if | 486:2 | `calendar_version` empty/non-canonical → refuse | `TestClosedBarRejectsMalformedIdentityStrings` (calendar version is empty) |
| B8 | if | 489:2 | `interval_ms` ≠ 60000 → refuse | `TestClosedBarRejectsIntervalOtherThanOneMinute` |
| B9 | if | 492:2 | `bar_label` ≠ `open_at` → refuse | `TestClosedBarRejectsUnknownEnumValues` /bar_label |
| B10 | if | 495:2 | `open_at_ms` zero or not on a whole minute → refuse | `TestClosedBarRejectsOpenAtThatIsNotOnTheMinute` |
| B11 | if | 498:2 | `finality` ≠ `successor_observed` → refuse | `TestClosedBarRejectsUnknownEnumValues` /finality |
| B12 | if | 501:2 | `closed` false → refuse (unfinished bar) | `TestClosedBarRejectsUnfinishedBar` |
| B13 | if | 505:2 | `source_observed_at_ms` < `open_at_ms + interval_ms` → refuse (future/unfinished) | `TestClosedBarRequiresObservationAfterTheBarClosed` (+59,999 ms refused, +60,000 accepted), `TestClosedBarRejectsBarFromTheFuture` |
| B14 | if | 511:2 | `marketclock.ParseMarket` error → refuse | not-applicable: unreachable — B2 already restricted `market` to `KR`/`US`, both accepted by `ParseMarket` (defensive) |
| B15 | if | 515:2 | `TradingDay` error → refuse | not-applicable: unreachable — `TradingDay` fails only for an unknown market or an unloadable zone; markets are pinned by B2/B14 and `time/tzdata` is embedded (defensive) |
| B16 | if | 518:2 | market-local day of `open_at_ms` ≠ session date → refuse | `TestClosedBarPayloadAloneRefusesABarFromAnotherCalendarDay` (header valid for 08-14, payload bar on 08-13/08-12/08-15 — the only fixture that reaches this payload-level check, since the header helper refuses first in `TestClosedBarRequiresTheSessionCalendarDay`); accept side `TestClosedBarAcceptsEveryBarOfTheMarketLocalSessionDay` |
| B17 | range | 522:2 | iterate open/high/low/close minors | every accepted bar |
| B18 | if | 526:3 | a price minor is 0 → refuse | `TestClosedBarRejectsImpossiblePriceOrdering` /zero low |
| B19 | if | 530:2 | `low_minor > high_minor` → refuse | `TestClosedBarRejectsImpossiblePriceOrdering` /low above high (refusal names `low_minor`, i.e. this rule fires before B20/B21) |
| B20 | if | 533:2 | `open_minor` outside [low, high] → refuse | `TestClosedBarRejectsImpossiblePriceOrdering` /open below low |
| B21 | if | 536:2 | `close_minor` outside [low, high] → refuse | `TestClosedBarRejectsImpossiblePriceOrdering` /close above high |
| B22 | range | 539:2 | iterate raw↔minor pairs (prices at market scale, volume at 0) | every accepted bar |
| B23 | if | 550:3 | `checkRawMinor` error (raw malformed / over-precise / leading zero / minor ≠ raw×scale) → refuse | `TestClosedBarRejectsMinorThatDisagreesWithRawDecimal`, `TestClosedBarRejectsOverPreciseRawForTheDeclaredScale`, `TestClosedBarRejectsSignedExponentOrPaddedRawDecimal`, `TestClosedBarRefusesLeadingZeroRawDecimal` |
| B24 | if | 554:2 | `revision` 0 → refuse | `TestClosedBarRejectsMalformedIdentityStrings` /revision is zero |
| B25 | if | 558:2 | `successor_open_at_ms` not on a whole minute → refuse | `TestClosedBarRequiresSuccessorOpenAt` /not on the minute |
| B26 | if | 561:2 | `successor_open_at_ms` < `open_at_ms + interval_ms` → refuse | `TestClosedBarRequiresSuccessorOpenAt` /at the bar itself, /before the bar, /zero |
| B27 | if | 564:2 | `successor_open_at_ms` > `source_observed_at_ms` → refuse | `TestClosedBarRequiresSuccessorOpenAt` /after the observation instant (equality accepted in the same test) |
| B28 | if | 567:2 | `source_response_digest` not `sha256:`+64 lower hex → refuse | `TestClosedBarRejectsMalformedIdentityStrings` (digest cases) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `marketMoney(bar.Market)` | 469 | single scale table (KR→KRW/0, US→USD/4); mutants 4→3 / 4→5 killed (review.md 2026-08-16 follow-up) |
| `canonicalStoredSymbolText`, `sessionDateFor`, `canonicalPayloadText`, `canonicalSHA256Digest` | 479, 482, 486, 567 | pure text predicates in the same file |
| `marketclock.ParseMarket`, `market.TradingDay(time.UnixMilli(...).UTC())` | 510, 514 | market-local calendar day (`internal/clock`, the package's only non-stdlib import) |
| `checkRawMinor(...)` → `minorFromRawDecimal` | 550 | integer-only recomputation; see `internal-strategyevidence--minorfromrawdecimal` |
| `payloadFieldError`, `formatUint`, `strconv.Quote` | each refusal | typed refusal text only |

## State mutations and fallbacks

- None. Locals only (`kind`, `currency/scale/known`, `sessionDate`, `market`, `barDay`, loop variables); no assignments outside the function, no defers, no goroutines (AST). No fallback: the first violated rule refuses the whole payload; nothing is coerced, rounded or defaulted.

## Safety conclusion

- This is the single value-level gate for `official_closed_bar_1m:v1` and is re-run on `NewEnvelope`, `Store.Append`, `scanEnvelope` replay and `SealBarSeries` read (through `validateTypedPayload` → `decodeClosedBar1mObject`). All 26 refusal branches except the two defensive ones (B14, B15) are pinned by named tests; both independent reviewers re-verified the calendar-day (B16), finality-minimum (B13), successor (B25–B27) and symbol (B5) rules at write and read (review.md 2026-08-17 rechecks). No order, sizing, auth or runtime surface is touched.
