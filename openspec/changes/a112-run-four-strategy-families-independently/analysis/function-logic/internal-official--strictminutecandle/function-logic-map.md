# Function Logic Map: `strictMinuteCandle`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `strictMinuteCandle(raw []byte, currency string) (RawMinuteCandle, time.Time, error)`
- Source range: `324:1`–`380:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 10, returns 9, calls 29, assignments 6, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- One candle object against the seven-key contract of decision 14: exactly `timestamp, openPrice, highPrice, lowPrice, closePrice, volume, currency`, every value a JSON **string** (a bare number refuses), non-empty, decimals at most 30 bytes (documented `maxLength`).
- Key count is checked before key lookup so an added broker field fails visibly instead of being dropped — chosen fail-closed behaviour, with the cost recorded in decision 23 (a benign field addition becomes an outage until the contract is re-measured).
- Numeric grammar is deliberately **not** validated here (decision 14): `strategyevidence` is the single decimal authority, and duplicating the grammar in a second package is exactly the duplication L1a's review flagged. This function only judges string-ness, emptiness and length.
- Currency must equal the market currency derived from the `market` argument (`KR→KRW`, `US→USD`), so a US page carrying `KRW` refuses.
- Timestamp must satisfy the grammar (numeric offset mandatory, optional 1–3 fractional digits), name an instant that exists, and — ruling 26 — start a minute. Ruling 26 exists so that one off-minute bar cannot poison the successor claim of the bar below it; the trade-off (that bar refuses the whole page) is recorded and accepted.

## Branches and early returns

Exact AST return nodes: `327, 332, 340, 345, 353, 358, 363, 368, 371`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 326:2 | the element is not an object → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `candle is not an object`) |
| B2 | if | 331:2 | the candle does not carry exactly seven keys → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `unknown candle key` = 8 keys, `missing candle key` = 6 keys) |
| B3 | range | 337:2 | pull each of the seven contract keys in fixed order | every accepting body test, e.g. `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`, `TestStrictMinuteCandlesAcceptsTheKoreanMarket` |
| B4 | if | 339:3 | a contract key is absent although the count is seven → `CANDLE_INVALID` | untested: reachable only with exactly seven keys of which one is renamed; the fixtures reach the count rule (B2) first — recorded gap |
| B5 | if | 344:3 | a value is not a non-empty JSON string → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `bare number price`, `bare number volume`, `null price`, `empty price string`) |
| B6 | range | 349:2 | walk the five decimal fields for the length bound | every accepting body test |
| B7 | if | 352:3 | a decimal longer than 30 bytes → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `over long decimal`, 31 digits) |
| B8 | if | 357:2 | the candle's currency is not the market currency → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `foreign currency`: `KRW` on a US page) |
| B9 | if | 362:2 | the timestamp fails the grammar or names an instant that does not exist → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `timestamp without an offset`, `timestamp with a zulu offset`, `timestamp that does not exist`) |
| B10 | if | 367:2 | ruling 26: the instant does not start a minute → `CANDLE_NOT_ON_MINUTE` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `instant with non-zero seconds`, `instant with a fraction`) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `strictMinuteObject(raw)` | 325 | pure extraction of the candle's keys and raw values; see `internal-official--strictminuteobject` |
| `len(fields)` against `len(strictMinuteCandleKeys)` | 331 | the seven-key rule |
| `strictMinuteString(value)` | 343 | string-ness and non-emptiness for every field |
| `len(values[key])` against `strictMinuteMaxDecimal` | 352 | documented 30-byte `maxLength` on the five decimals |
| `strictMinuteInstant(values["timestamp"])` | 361 | grammar then existence (a syntactically valid 30 February is caught by the parse) |
| `instant.Second()`, `instant.Nanosecond()` | 367 | ruling 26 minute alignment |

## State mutations and fallbacks

- Locals only (6 AST assignments): `fields`, `values`, `value`/`found`, `text`/`err`, `instant`. No package state, no I/O, no goroutines, no defers, no clock read.
- No fallback: every refusal returns the zero `RawMinuteCandle` and the zero instant, so a partially decoded candle can never escape. The raw strings are carried through verbatim — no normalisation, no re-formatting, no numeric conversion.

## Safety conclusion

- The single place where a stored bar's seven fields are bound to the bytes the broker sent. Refusing rather than coercing is what keeps `strategyevidence`'s decimal authority meaningful: it receives exactly the received strings or nothing.
- Ruling 26 is the safety-relevant addition of the fix round: minute alignment is checked here, at the earliest point, so an off-minute bar cannot reach the producer and corrupt its neighbour's `successor_open_at_ms`. The recorded cost is that one such bar refuses the whole page.
- One untested-but-reachable branch (B4, a seven-key candle with a renamed key) is recorded above rather than claimed as covered; it is a strictly narrower path than B2 and fails closed the same way.
