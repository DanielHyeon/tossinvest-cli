# Function Logic Map: `strictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `strictMinuteCandles(raw []byte, count int, currency string, beforeInstant time.Time) ([]RawMinuteCandle, []time.Time, error)`
- Source range: `290:1`–`322:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 7, returns 7, calls 25, assignments 7, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Input is the raw bytes of `result.candles` (already proven well-formed by `strictMinuteCheckJSON`), the requested `count`, the market currency and the optional inclusive upper bound; output is the decoded candles in received order plus their parsed instants.
- Array-level contract of decision 14: the value must be a JSON array; the page may carry at most `count` candles (and `count` itself is bounded at 200 by the caller); instants must be strictly descending and unique; when a `before` was sent, every instant must be ≤ it because the documented bound is inclusive.
- The instants slice is returned so the cursor rule can compare against the page's oldest bar without re-parsing.
- Two-step array check: a cheap `[` sniff on the trimmed bytes names the failure precisely ("candles is not a JSON array") before `json.Unmarshal` is asked to shape it, so `null` and `{}` refuse with the same reason.

## Branches and early returns

Exact AST return nodes: `293, 297, 300, 308, 311, 315, 321`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 292:2 | the value is empty or does not open with `[` → `RESULT_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `candles is not an array`, `candles is null`) |
| B2 | if | 296:2 | unmarshalling into `[]json.RawMessage` failed → `RESULT_INVALID` | not-applicable: `strictMinuteCheckJSON` has already walked the whole body, so once B1 has proved the value opens as an array the element split cannot fail; defensive |
| B3 | if | 299:2 | the page carries more candles than were asked for → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMoreCandlesThanRequested` (3 for count 2 refused, 3 for count 3 accepted) |
| B4 | range | 305:2 | walk the rows newest→oldest, decoding and ordering each | every accepting body test, e.g. `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` (2 candles), `TestStrictMinuteCandlesAcceptsTheKoreanMarket` (1 candle) |
| B5 | if | 307:3 | a candle refused → propagate unchanged (key set, string-ness, length, currency, timestamp, minute alignment) | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `unknown candle key`, `missing candle key`, `candle is not an object`, `bare number price`, `bare number volume`, `null price`, `empty price string`, `over long decimal`, `foreign currency`, `timestamp without an offset`, `timestamp with a zulu offset`, `timestamp that does not exist`, `instant with non-zero seconds`, `instant with a fraction`) |
| B6 | if | 310:3 | this instant is not strictly older than the previous one → `CANDLE_ORDER_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `ascending instants`, `duplicate instants`) |
| B7 | if | 314:3 | this instant is newer than the inclusive `before` bound → `CANDLE_ORDER_INVALID` | `TestStrictMinuteCandlesRefusesInstantsNewerThanBefore` (one minute past the bound refused; a bar exactly at the bound accepted) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `bytes.TrimLeft(raw, " \t\r\n")` | 291 | whitespace-tolerant array sniff |
| `json.Unmarshal(raw, &rows)` | 296 | element split into `[]json.RawMessage` (allocation precedes the count bound — recorded residual, bounded by the 2 MiB body cap) |
| `strictMinuteCandle(row, currency)` | 306 | the per-candle contract; see `internal-official--strictminutecandle` |
| `instant.Before(instants[index-1])` | 310 | strict descending order |
| `beforeInstant.IsZero()`, `instant.After(beforeInstant)` | 314 | inclusive upper bound applied only when a `before` was sent |

## State mutations and fallbacks

- Locals only (7 AST assignments): `trimmed`, `rows`, `candles`, `instants`, `candle`/`instant`/`err`. No package state, no I/O, no goroutines, no defers, no clock read.
- No fallback: any refusal discards the whole page (`return nil, nil, err`). Nothing is truncated to `count` and no candle is skipped — a page that carries too many is refused rather than trimmed, because a trimmed page would silently change which bar is "newest".

## Safety conclusion

- Array-level fail-closed layer. Its ordering rule is what lets the producer treat `bars[i-1]` as the successor of `bars[i]`: without strict descending uniqueness, decision 6's "never admit the newest bar" would bind the wrong minute.
- The inclusive-bound rule keeps the process clock honest as a *query* bound rather than an authority — a page can never contain a bar newer than the instant the caller asked about.
- Recorded residual (review.md 2026-08-17): `[]json.RawMessage` is allocated before the count bound is enforced; the exposure is bounded by the 2 MiB post-read body cap (≈20× the measured page). One defensive branch (B2) is unreachable behind the whole-body walk.
