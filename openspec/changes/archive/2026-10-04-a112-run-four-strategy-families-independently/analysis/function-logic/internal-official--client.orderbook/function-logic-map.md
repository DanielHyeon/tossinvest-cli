# Function Logic Map: `Client.Orderbook`

- Source: `internal/official/market_reads.go`
- Source SHA-256: `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea` (current worktree; verified with `sha256sum` 2026-08-18, equal to `source_sha256` in `ast.json`)
- Signature: `(c *Client) Orderbook(ctx context.Context, symbol string) (domain.OrderBook, error)` (`ast.json`: `Client.Orderbook(params=2, results=2)`)
- Source range: `209:1`-`217:2`
- AST counts: branches 1, returns 2, calls 3, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief. Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbol` | any string | caller | not validated here; sent as the sole query parameter |
| request | `GET /api/v1/orderbook?symbol=<symbol>` | `q.Set` at 211:2 | no depth/level-count parameter exists in this call - the ladder length is whatever the broker returns |

- There is no `count`, `depth` or `levels` parameter. Level count is therefore a property of the response, per market, and the L1c reader must treat it as data rather than a constant (KR and US returned different depths in the 2026-08-18 probe).

## Branches and early returns

Exact AST return nodes: `214, 216`.

- B1 (`if` at 213:2) - `c.get(ctx, "/api/v1/orderbook", q, &raw)` returned an error. Returns the zero `domain.OrderBook` and that error at 214:3. `c.get` is the ordinary production path (token adopt-before-buy, <=2 refresh on 401, `classifyStatus`) documented in the `internal-official--client.send` and `--client.dorequest` bundles; the raw bytes are consumed by `unwrapAndDecode` and are not retained.
- Fall-through - returns `adaptOrderbook(symbol, raw)` at 216:2 with a nil error.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `q.Set` | 211:2 |
| `c.get` | 213:12 |
| `adaptOrderbook` | 216:9 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `q.Set` (211:2) | build the single query parameter | none | `ast.json` call |
| `c.get` (213:12) | perform the authenticated GET and decode `result` into `apiOrderbook` | ordinary `send`: token refresh loop, non-2xx -> `classifyStatus`, permissive `encoding/json` (unknown keys ignored, duplicates last-wins) | `ast.json` call; `internal-official--client.send`, `--unwrapanddecode` bundles |
| `adaptOrderbook` (216:9) | lossy projection to `domain.OrderBook` | none | `ast.json` call; `internal-official--adaptorderbook` bundle |

## State mutations and fallbacks

- No package state. The lossiness is entirely downstream: `c.get` discards the bytes and `adaptOrderbook` discards currency and timestamp, so this path cannot mint evidence. L1c's reader must observe the response bytes on the way past, exactly as `StrictMinuteCandles` does for candles.

## Safety conclusion

- Safe edit boundary: a112 does not edit this function. It is a two-line request/adapt pair on the shared official client; L1c adds a sibling strict reader (the `StrictMinuteCandles` precedent) and leaves this caller untouched.
- High-risk impact: no by role, yes by adjacency - it runs on the production token path (`c.get` -> `send`, <=2 refresh on 401), which is shared with the live containers and with the neighbouring product's credential (token-war memory). Any L1c run therefore consumes the same quota and may refresh the same cached token; that is why probes stay human-approved and one-shot.
- Untested branch: B1 (the `c.get` error path). The success path is pinned by `TestOrderbookIntegration`; the package suite is green (`go test ./internal/official -count=1`, 351 tests, exit 0, 2026-08-18).
