# Function Logic Map: `strictMinuteCursor`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `strictMinuteCursor(raw []byte, present bool, instants []time.Time) (bool, string, error)`
- Source range: `384:1`–`409:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 6, returns 7, calls 16, assignments 3, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- The cursor tri-state that the standard decoder cannot express (the file's opening rationale): a JSON **string** means "there is another page", `null` means "this was the last page", and **absent** is neither — it is a contract break. The standard decoder flattens all three to `""`, which is why this reader exists.
- Decision 14, kept by decision 23 as a recorded residual: an absent `nextBefore` is refused although the documented schema lists only `candles` as required, because a terminal page has never been measured. It stays fail-closed until one is.
- Loop guard: a string cursor must be strictly older than the page's oldest bar. The measured cursor is "last bar − 1 minute" and is an inclusive bound for the next request, so a cursor at or after the oldest bar would let a crawl stand still.
- Ruling 26 does **not** apply here: a cursor is a bound, not a bar, so sub-minute cursors are accepted deliberately (the measured value happened to sit on a minute; the contract requires only strict ordering).
- With an empty page the ordering rule has nothing to compare against, so an empty page carrying a string cursor is a valid page — the producer, not the reader, decides to stop on it.

## Branches and early returns

Exact AST return nodes: `386, 391, 394, 398, 402, 405, 408`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 385:2 | the `nextBefore` key is absent → `CURSOR_INVALID` ("absent is not terminal") | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `cursor replaced by another key`, the two-key body that the key-count rule cannot catch) |
| B2 | if | 390:2 | the value is `null` → terminal page, empty cursor | `TestStrictMinuteCandlesAcceptsTheKoreanMarket`, `TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys`, `TestStrictMinuteCandlesAcceptsNestingUpToTheBound` (all assert `page.Terminal`) |
| B3 | if | 393:2 | the value is neither `null` nor a JSON string → `CURSOR_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `cursor is a number`, `cursor is an object`, `cursor is an array`, `cursor is a boolean`) |
| B4 | if | 397:2 | the string is empty (or not extractable) → `CURSOR_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `cursor is an empty string`) |
| B5 | if | 401:2 | the cursor fails the timestamp grammar or names an instant that does not exist → `CURSOR_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `cursor breaks the timestamp grammar`, a `Z` offset) |
| B6 | if | 404:2 | the cursor is not strictly older than this page's oldest bar → `CURSOR_INVALID` | taken: `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `cursor is not older than the oldest bar`, `cursor is newer than the oldest bar`); untaken with an empty page: `TestStrictMinuteCandlesAcceptsAnEmptyPageThatStillCarriesACursor`; sub-minute cursor accepted: `TestStrictMinuteCandlesAcceptsASubMinuteCursor` |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `bytes.TrimSpace(raw)`, `bytes.Equal(trimmed, []byte("null"))` | 389–390 | the `null` arm is decided on the raw bytes, not on a decoded `any` |
| `strictMinuteString(trimmed)` | 396 | string-ness and non-emptiness |
| `strictMinuteInstant(value)` | 400 | the same grammar the request's `before` had to satisfy |
| `cursor.Before(instants[len(instants)-1])` | 404 | the loop guard against the page's oldest bar |
| `strictMinuteRefuse(StrictReasonCursor, …)` ×5 | 386, 394, 398, 402, 405 | one reason label for every cursor break, asserted by the reader table |

## State mutations and fallbacks

- Locals only (3 AST assignments): `trimmed`, `value`, `cursor`. No package state, no I/O, no goroutines, no defers, no clock read.
- No fallback: an unusable cursor never degrades to "terminal" and never degrades to "empty string". That degradation is precisely the failure mode this function was written to remove — a page wrongly read as terminal would silently truncate a session's evidence.

## Safety conclusion

- Small function, high leverage: it decides whether a crawl continues, and a wrong "terminal" would end a session's evidence early without any error. Every ambiguous shape refuses instead.
- The producer adds a second, independent guard on top of this one: `PollClosedBars` B27 requires the new cursor to be strictly older than the bound the request carried (D4), so a reader that returned a stale-but-well-formed cursor still cannot spin the crawl.
- Recorded residuals (review.md 2026-08-17): the absent-cursor refusal is stricter than the documented schema (fail-closed until a terminal page is measured); the KR pagination contract is documented but not measured — M-B measured US only, so the first KR full-session crawl under L5 is a human-observed run.
