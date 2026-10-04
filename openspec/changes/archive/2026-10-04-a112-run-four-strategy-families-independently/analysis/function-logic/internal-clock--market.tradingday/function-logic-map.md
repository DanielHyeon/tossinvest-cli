# Function Logic Map: `Market.TradingDay`

- Source: `internal/clock/market.go`
- Source SHA-256: `85264b6e9b4e6b13ddee690ca62e0d30cbd73efeb6926e626abc3d20c622f7a9` (current worktree; verified with `sha256sum` 2026-08-17, equal to `source_sha256` in `ast.json`)
- Signature: `(m Market) TradingDay(t time.Time) (string, error)` (`ast.json`: `Market.TradingDay(params=1, results=2)`)
- Source range: `158:1`–`164:2`
- AST counts: branches 1, returns 2, calls 3, defers 0, go statements 0 (`ast.json` generated 2026-08-17 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

- The trading-day boundary for the whole product: one UTC instant belongs to different trading days in Seoul and New York, and journal records label the day per market rather than per host timezone (market.go:154–157). The L1b brief cites it because a US bar's calendar day must be the market-local one, not the UTC one.
- `m` must be a market `Location()` knows. `Location` refuses anything outside the `tzName` table with `ErrUnknownMarket` and also fails closed if the one-time zone load failed (market.go:79–92), so this function inherits a fail-closed market check rather than defaulting to an exchange.
- `t` is unconstrained — any instant, any zone, including the zero time. There is no session, holiday or weekend judgement here: the answer is purely the market-local calendar date.
- Output shape is fixed at `2006-01-02`. Callers compare these strings for equality (`SameTradingDay`, market.go:168–), so the layout is part of the contract.

## Branches and early returns

Exact AST return nodes: `161, 163`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 160:2 | `m.Location()` failed → return `("", err)`; an empty day is never presented as a real one | `TestUnknownMarketFailsClosed` (asserts `TradingDay` returns `ErrUnknownMarket` for the market `nasdaq`, alongside the other market-scoped judgements) |

Fall-through (line 163): `t.In(loc).Format("2006-01-02")` with a nil error — covered by `TestTradingDayBoundary`, which pins one UTC instant onto two different KR days (13:30Z → 2026-03-30, 20:00Z → 2026-03-31) and the same US day (both 2026-03-30).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `m.Location` | 159:14 |
| `Format` | 163:9 |
| `t.In` | 163:9 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `m.Location()` | 159 | the market's IANA zone, loaded once behind `locOnce`; refuses unknown markets and a failed zone load (market.go:79–92) |
| `t.In(loc)` | 163 | reinterpret the instant in the exchange's zone; no arithmetic, so DST transitions are handled by the zone database (`TestUTCOffsetAcrossDST`, `TestRegularSessionDSTTable`) |
| `Format("2006-01-02")` | 163 | the fixed day layout callers compare on |

## State mutations and fallbacks

- No mutation. Value receiver, one local (`loc`, a single AST assignment), no package state written. `Location` may trigger the process-wide `locOnce` zone load, which is a read-only initialisation shared by every market helper.
- No fallback: an unknown market yields the empty string *and* an error together, so a caller that ignores the error gets a value that cannot be mistaken for a date.

## Safety conclusion

- Safe edit boundary: the layout string and the fail-closed error are the contract. Substituting a default location, or returning a formatted host-local date when `Location` fails, would silently relabel journal and evidence rows across the day boundary.
- High-risk impact: no direct order, stop or sizing effect, but it labels the day that ledger, journal and evidence rows are keyed by, so a wrong answer misfiles records rather than mispricing a trade. It is single-branch and fail-closed.
- Untested branches: none. The single branch and the fall-through are both covered. Package suite green (`go test ./internal/clock -count=1`, 2026-08-17).
