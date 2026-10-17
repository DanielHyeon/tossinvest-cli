# Function Logic Map: `strictMinuteCandle`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `strictMinuteCandle(raw []byte, currency string) (RawMinuteCandle, time.Time, error)`
- Source range: `335:1`–`392:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 10, returns 9, calls 29, assignments 6, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- One candle object against the seven-key contract of decision 14: exactly `timestamp, openPrice, highPrice, lowPrice, closePrice, volume, currency`, every value a JSON **string** (a bare number refuses), non-empty, decimals at most 30 bytes (documented `maxLength`).
- Key count is checked before key lookup so an added broker field fails visibly instead of being dropped — chosen fail-closed behaviour, with the cost recorded in decision 23 (a benign field addition becomes an outage until the contract is re-measured).
- Numeric grammar is deliberately **not** validated here (decision 14): `strategyevidence` is the single decimal authority, and duplicating the grammar in a second package is exactly the duplication L1a's review flagged. This function only judges string-ness, emptiness and length.
- Currency must equal the market currency derived from the `market` argument (`KR→KRW`, `US→USD`), so a US page carrying `KRW` refuses.
- **Decision 30 (2026-08-18), documentation-only.** `timestamp` is the bar's **close** instant, not its open (US probe 03:29 KST; review.md); the source comment at 376–378 now says so. The value is carried through verbatim as a string and this function converts nothing — the producer's `adoptPage` subtracts one interval. Ruling 26's alignment rule is therefore about the *label* landing on a minute boundary; the producer re-asserts alignment on the converted side.
- Timestamp must satisfy the grammar (numeric offset mandatory, optional 1–3 fractional digits), name an instant that exists, and — ruling 26 — land on a minute boundary. Ruling 26 exists so that one off-minute bar cannot poison the successor claim of the bar below it; the trade-off (that bar refuses the whole page) is recorded and accepted.

## Branches and early returns

Exact AST return nodes: `338`, `343`, `351`, `356`, `364`, `369`, `374`, `380`, `383`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 337:2 | the element is not an object → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `candle is not an object`) |
| B2 | if | 342:2 | the candle does not carry exactly seven keys → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `unknown candle key` = 8 keys, `missing candle key` = 6 keys) |
| B3 | range | 348:2 | pull each of the seven contract keys in fixed order | every accepting body test, e.g. `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`, `TestStrictMinuteCandlesAcceptsTheKoreanMarket` |
| B4 | if | 350:3 | a contract key is absent although the count is seven → `CANDLE_INVALID` | untested: reachable only with exactly seven keys of which one is renamed; the fixtures reach the count rule (B2) first — recorded gap |
| B5 | if | 355:3 | a value is not a non-empty JSON string → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `bare number price`, `bare number volume`, `null price`, `empty price string`) |
| B6 | range | 360:2 | walk the five decimal fields for the length bound | every accepting body test |
| B7 | if | 363:3 | a decimal longer than 30 bytes → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `over long decimal`, 31 digits) |
| B8 | if | 368:2 | the candle's currency is not the market currency → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `foreign currency`: `KRW` on a US page) |
| B9 | if | 373:2 | the timestamp fails the grammar or names an instant that does not exist → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `timestamp without an offset`, `timestamp with a zulu offset`, `timestamp that does not exist`) |
| B10 | if | 379:2 | ruling 26: the instant does not start a minute → `CANDLE_NOT_ON_MINUTE` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `instant with non-zero seconds`, `instant with a fraction`) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strictMinuteObject` | 336:17 |
| `strictMinuteRefuse` | 338:42 |
| `err.Error` | 338:81 |
| `len` | 342:5 |
| `len` | 342:20 |
| `strictMinuteRefuse` | 343:42 |
| `strconv.Itoa` | 344:22 |
| `len` | 344:35 |
| `strconv.Itoa` | 345:5 |
| `len` | 345:18 |
| `make` | 347:12 |
| `len` | 347:36 |
| `strictMinuteRefuse` | 351:43 |
| `strconv.Quote` | 352:22 |
| `strictMinuteString` | 354:16 |
| `strictMinuteRefuse` | 356:43 |
| `err.Error` | 356:91 |
| `len` | 363:6 |
| `strictMinuteRefuse` | 364:43 |
| `strconv.Itoa` | 365:28 |
| `strictMinuteRefuse` | 369:42 |
| `strconv.Quote` | 370:16 |
| `strictMinuteInstant` | 372:18 |
| `strictMinuteRefuse` | 374:42 |
| `err.Error` | 374:81 |
| `instant.Second` | 379:5 |
| `instant.Nanosecond` | 379:30 |
| `strictMinuteRefuse` | 380:42 |
| `strconv.Quote` | 381:17 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `strictMinuteObject(raw)` | 336 | pure extraction of the candle's keys and raw values; see `internal-official--strictminuteobject` |
| `len(fields)` against `len(strictMinuteCandleKeys)` | 342 | the seven-key rule |
| `strictMinuteString(value)` | 354 | string-ness and non-emptiness for every field |
| `len(values[key])` against `strictMinuteMaxDecimal` | 363 | documented 30-byte `maxLength` on the five decimals |
| `strictMinuteInstant(values["timestamp"])` | 372 | grammar then existence (a syntactically valid 30 February is caught by the parse) |
| `instant.Second()`, `instant.Nanosecond()` | 379 | ruling 26 minute alignment |

## State mutations and fallbacks

- Locals only (6 AST assignments): `fields`, `values`, `value`/`found`, `text`/`err`, `instant`. No package state, no I/O, no goroutines, no defers, no clock read.
- No fallback: every refusal returns the zero `RawMinuteCandle` and the zero instant, so a partially decoded candle can never escape. The raw strings are carried through verbatim — no normalisation, no re-formatting, no numeric conversion.

## Safety conclusion

- The single place where a stored bar's seven fields are bound to the bytes the broker sent. Refusing rather than coercing is what keeps `strategyevidence`'s decimal authority meaningful: it receives exactly the received strings or nothing.
- Ruling 26 is the safety-relevant addition of the fix round: minute alignment is checked here, at the earliest point, so an off-minute bar cannot reach the producer and corrupt its neighbour's `successor_open_at_ms`. The recorded cost is that one such bar refuses the whole page.
- One untested-but-reachable branch (B4, a seven-key candle with a renamed key) is recorded above rather than claimed as covered; it is a strictly narrower path than B2 and fails closed the same way.
