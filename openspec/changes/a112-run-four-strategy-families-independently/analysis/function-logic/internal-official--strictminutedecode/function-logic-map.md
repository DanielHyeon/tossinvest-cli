# Function Logic Map: `strictMinuteDecode`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `strictMinuteDecode(body []byte, count int, currency string, beforeInstant time.Time) ([]RawMinuteCandle, bool, string, error)`
- Source range: `243:1`–`288:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 10, returns 11, calls 22, assignments 8, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- The body-contract gate of decision 14. Order matters and is deliberate: size cap → UTF-8 validity → one whole-body well-formedness walk (duplicate keys at any depth, trailing value, depth bound) → envelope object → `result` object → key-set → candles → cursor. Every later step therefore works on bytes already proven well-formed.
- Size: bodies above 2 MiB are refused. The cap is applied **after** the read because `doRequest` reads uncapped — a transport-wide residual accepted for this lot.
- UTF-8 first, deliberately: the standard decoder replaces broken bytes with U+FFFD, which would turn a byte the broker never sent into evidence. Ruling 29 added a test that hides the bad byte inside `openPrice` rather than in the cursor, so the rule is not masked by a later check.
- Envelope: unknown sibling keys are ignored (M-B0 precedent — `traceId` and friends), but `result` must be present and must be an object.
- `result` key-set is exactly `{candles, nextBefore}`. `candles` presence is checked first so its absence names the right key; the count check then rejects an unknown extra key. A body with two keys where the second is not `nextBefore` passes both and is caught by the cursor rule alone — the review's declared reason for keeping "absent is not terminal" as a separate refusal.
- Numeric grammar is **not** validated here: `strategyevidence` is the single numeric authority (decision 14). This function only proves shape, string-ness, length and time.

## Branches and early returns

Exact AST return nodes: `245, 251, 254, 258, 263, 267, 272, 275, 281, 285, 287`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 244:2 | body above the 2 MiB cap → `BODY_TOO_LARGE` (post-read cap) | `TestStrictMinuteCandlesRefusesABodyAboveTwoMebibytes` |
| B2 | if | 250:2 | the body is not valid UTF-8 → `BODY_NOT_UTF8`, before any decoding | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `invalid utf-8 inside a price`), `TestStrictMinuteCandlesRefusesInvalidUTF8` (bad bytes in the cursor) |
| B3 | if | 253:2 | the whole-body walk refused (duplicate key at any depth, trailing value, depth bound, unreadable stream) → `BODY_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `duplicate key in the envelope`, `duplicate key in the result`, `duplicate key in a candle`, `duplicate key inside an ignored sibling`, `trailing json value`, `empty body`, `seventeen levels of nesting`) |
| B4 | if | 257:2 | the body is well-formed JSON but not an object → `BODY_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `body is not an object`) |
| B5 | if | 262:2 | the envelope has no `result` key → `RESULT_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `result absent`) |
| B6 | if | 266:2 | `result` is not an object → `RESULT_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `result is not an object`) |
| B7 | if | 271:2 | `result` has no `candles` key → `RESULT_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `candles absent`) |
| B8 | if | 274:2 | `result` does not carry exactly two keys → `RESULT_INVALID` (unknown key, or `nextBefore` simply missing) | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `unknown result key`, `cursor absent`) |
| B9 | if | 280:2 | the candle array refused → propagate the typed refusal unchanged | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `candles is not an array`, `candles is null`, `unknown candle key`, `missing candle key`, `bare number price`, `foreign currency`, `ascending instants`, `instant with non-zero seconds`, …), `TestStrictMinuteCandlesRefusesMoreCandlesThanRequested`, `TestStrictMinuteCandlesRefusesInstantsNewerThanBefore` |
| B10 | if | 284:2 | the cursor refused → propagate the typed refusal unchanged | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `cursor replaced by another key`, `cursor is an empty string`, `cursor is a number`, `cursor is an object`, `cursor is an array`, `cursor is a boolean`, `cursor breaks the timestamp grammar`, `cursor is not older than the oldest bar`, `cursor is newer than the oldest bar`) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `len(body)` against `strictMinuteMaxBody` | 244 | 2 MiB post-read cap |
| `utf8.Valid(body)` | 250 | ruling 29's UTF-8 rule, tested through a price field |
| `strictMinuteCheckJSON(body)` | 253 | single whole-body authority for duplicate keys, trailing values and depth; see `internal-official--strictminutewalk` |
| `strictMinuteObject(body)`, `strictMinuteObject(resultRaw)` | 256, 265 | pure extractors over already-validated bytes; see `internal-official--strictminuteobject` |
| `strictMinuteCandles(candlesRaw, count, currency, beforeInstant)` | 279 | array, count bound, per-candle contract, ordering and the inclusive `before` bound |
| `strictMinuteCursor(cursorRaw, hasCursor, instants)` | 283 | the cursor tri-state (string / `null` / refuse) |

## State mutations and fallbacks

- Locals only (8 AST assignments): `envelope`, `resultRaw`/`found`, `result`, `candlesRaw`/`hasCandles`, `cursorRaw`/`hasCursor`, `candles`/`instants`, `terminal`/`cursor`. No client or package state, no I/O, no goroutines, no defers, no clock read.
- No fallback anywhere: every refusal returns `nil, false, "", err`, so a caller cannot receive a partially decoded page. Unknown **envelope** siblings are the single tolerated deviation and they are ignored, never surfaced.

## Safety conclusion

- This is where "the broker changed shape" becomes a visible failure rather than a silent coercion, which is the point of the whole reader (decision 14). Every refusal is typed, so the producer and later lots can tell a contract break from a transport failure.
- Deliberate availability trade: an unknown key added benignly by the broker fails the read until the contract is re-measured and widened (decision 23), and ruling 26's minute-alignment rule can make one off-minute bar refuse a whole page. Both are recorded, chosen, and reviewed.
- Recorded residuals (review.md 2026-08-17): the 2 MiB cap is applied after an uncapped `doRequest` read; `strictMinuteCandles` allocates `[]json.RawMessage` before the count bound, inside that cap; an absent `nextBefore` is refused although the documented schema lists only `candles` as required.
