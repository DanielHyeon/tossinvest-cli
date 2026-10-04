# Function Logic Map: `strictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `strictMinuteCandles(raw []byte, count int, currency string, beforeInstant time.Time) ([]RawMinuteCandle, []time.Time, error)`
- Source range: `301:1`–`333:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 7, returns 7, calls 25, assignments 7, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Input is the raw bytes of `result.candles` (already proven well-formed by `strictMinuteCheckJSON`), the requested `count`, the market currency and the optional inclusive upper bound; output is the decoded candles in received order plus their parsed instants.
- Array-level contract of decision 14: the value must be a JSON array; the page may carry at most `count` candles (and `count` itself is bounded at 200 by the caller); instants must be strictly descending and unique; when a `before` was sent, every instant must be ≤ it because the documented bound is inclusive.
- The instants slice is returned so the cursor rule can compare against the page's oldest bar without re-parsing.
- **Decision 30 (2026-08-18), documentation-only.** Every instant handled here is a bar **close** label (US probe 03:29 KST; review.md), and so is `beforeInstant`. "Strictly descending" and "≤ the inclusive `before`" therefore compare like with like and did not change; because `before` bounds the close, a request with `before = now` already excludes the still-forming bar. No conversion happens in this package.
- Two-step array check: a cheap `[` sniff on the trimmed bytes names the failure precisely ("candles is not a JSON array") before `json.Unmarshal` is asked to shape it, so `null` and `{}` refuse with the same reason.

## Branches and early returns

Exact AST return nodes: `304`, `308`, `311`, `319`, `322`, `326`, `332`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 303:2 | the value is empty or does not open with `[` → `RESULT_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `candles is not an array`, `candles is null`) |
| B2 | if | 307:2 | unmarshalling into `[]json.RawMessage` failed → `RESULT_INVALID` | not-applicable: `strictMinuteCheckJSON` has already walked the whole body, so once B1 has proved the value opens as an array the element split cannot fail; defensive |
| B3 | if | 310:2 | the page carries more candles than were asked for → `CANDLE_INVALID` | `TestStrictMinuteCandlesRefusesMoreCandlesThanRequested` (3 for count 2 refused, 3 for count 3 accepted) |
| B4 | range | 316:2 | walk the rows newest→oldest, decoding and ordering each | every accepting body test, e.g. `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` (2 candles), `TestStrictMinuteCandlesAcceptsTheKoreanMarket` (1 candle) |
| B5 | if | 318:3 | a candle refused → propagate unchanged (key set, string-ness, length, currency, timestamp, minute alignment) | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `unknown candle key`, `missing candle key`, `candle is not an object`, `bare number price`, `bare number volume`, `null price`, `empty price string`, `over long decimal`, `foreign currency`, `timestamp without an offset`, `timestamp with a zulu offset`, `timestamp that does not exist`, `instant with non-zero seconds`, `instant with a fraction`) |
| B6 | if | 321:3 | this instant is not strictly older than the previous one → `CANDLE_ORDER_INVALID` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `ascending instants`, `duplicate instants`) |
| B7 | if | 325:3 | this instant is newer than the inclusive `before` bound → `CANDLE_ORDER_INVALID` | `TestStrictMinuteCandlesRefusesInstantsNewerThanBefore` (one minute past the bound refused; a bar exactly at the bound accepted) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `bytes.TrimLeft` | 302:13 |
| `len` | 303:5 |
| `strictMinuteRefuse` | 304:20 |
| `json.Unmarshal` | 307:12 |
| `strictMinuteRefuse` | 308:20 |
| `err.Error` | 308:71 |
| `len` | 310:5 |
| `strictMinuteRefuse` | 311:20 |
| `strconv.Itoa` | 312:20 |
| `len` | 312:33 |
| `strconv.Itoa` | 312:71 |
| `make` | 314:13 |
| `len` | 314:40 |
| `make` | 315:14 |
| `len` | 315:35 |
| `strictMinuteCandle` | 317:27 |
| `instant.Before` | 321:20 |
| `strictMinuteRefuse` | 322:21 |
| `strconv.Itoa` | 323:15 |
| `beforeInstant.IsZero` | 325:7 |
| `instant.After` | 325:33 |
| `strictMinuteRefuse` | 326:21 |
| `strconv.Itoa` | 327:15 |
| `append` | 329:13 |
| `append` | 330:14 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `bytes.TrimLeft(raw, " \t\r\n")` | 302 | whitespace-tolerant array sniff |
| `json.Unmarshal(raw, &rows)` | 307 | element split into `[]json.RawMessage` (allocation precedes the count bound — recorded residual, bounded by the 2 MiB body cap) |
| `strictMinuteCandle(row, currency)` | 317 | the per-candle contract; see `internal-official--strictminutecandle` |
| `instant.Before(instants[index-1])` | 321 | strict descending order |
| `beforeInstant.IsZero()`, `instant.After(beforeInstant)` | 325 | inclusive upper bound applied only when a `before` was sent |

## State mutations and fallbacks

- Locals only (7 AST assignments): `trimmed`, `rows`, `candles`, `instants`, `candle`/`instant`/`err`. No package state, no I/O, no goroutines, no defers, no clock read.
- No fallback: any refusal discards the whole page (`return nil, nil, err`). Nothing is truncated to `count` and no candle is skipped — a page that carries too many is refused rather than trimmed, because a trimmed page would silently change which bar is "newest".

## Safety conclusion

- Array-level fail-closed layer. Its ordering rule is what lets the producer treat `bars[i-1]` as the successor of `bars[i]`: without strict descending uniqueness, decision 6's "never admit the newest bar" would bind the wrong minute.
- The inclusive-bound rule keeps the process clock honest as a *query* bound rather than an authority — a page can never contain a bar newer than the instant the caller asked about.
- Recorded residual (review.md 2026-08-17): `[]json.RawMessage` is allocated before the count bound is enforced; the exposure is bounded by the 2 MiB post-read body cap (≈20× the measured page). One defensive branch (B2) is unreachable behind the whole-body walk.
