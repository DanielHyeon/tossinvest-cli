# Function Logic Map: `Client.Prices`

- Source: `internal/official/market_reads.go`
- Source SHA-256: `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea` (current worktree; verified with `sha256sum` 2026-08-18, equal to `source_sha256` in `ast.json`)
- Signature: `(c *Client) Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)` (`ast.json`: `Client.Prices(params=2, results=2)`)
- Source range: `142:1`-`150:2`
- AST counts: branches 1, returns 2, calls 4, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief (the `last` half of the quote seal). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbols` | up to 200 per the documented API limit | caller | not validated or capped here |
| request | `GET /api/v1/prices?symbols=<comma-joined>` | `strings.Join` + `q.Set` at 144:2 | a symbol containing a comma would silently split; no escaping is applied |
| response | `[]apiPrice` decoded from `result` | `c.get` at 146:2 | row count and row order are not checked against `symbols` |

- This is the only official read that yields a last price. The L1c quote seal needs `bid`, `ask` and `last` (`breakoutlane.QuoteSealInput`), so the seal spans two endpoints: this one and `/api/v1/orderbook`.

## Branches and early returns

Exact AST return nodes: `147, 149`.

- B1 (`if` at 146:2) - `c.get(ctx, "/api/v1/prices", q, &raw)` returned an error; returns `nil` and the error at 147:3.
- Fall-through - returns `adaptPrices(raw)` at 149:2 with a nil error, including when `raw` is empty: an unknown symbol yields an empty slice, not an error. `hybrid.Client.GetQuote` (`internal/hybrid/client.go:117-129`) is what converts that empty slice into `"hybrid: no quote found for %s"`, and only for the single-symbol case.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `q.Set` | 144:2 |
| `strings.Join` | 144:19 |
| `c.get` | 146:12 |
| `adaptPrices` | 149:9 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Join` / `q.Set` (144) | build the `symbols` parameter | none | `ast.json` calls |
| `c.get` (146:12) | authenticated GET, decode `result` into `[]apiPrice` | ordinary `send` (see `internal-official--client.send`) | `ast.json` call |
| `adaptPrices` (149:9) | lossy projection to `[]domain.Quote` | none | `ast.json` call; `internal-official--adaptprices` bundle |

## State mutations and fallbacks

- No package state. The evidence-relevant fallback is the silent empty result at the fall-through return: absence of a price is not an error on this path, which is precisely the shape a fail-closed producer must invert.

## Safety conclusion

- Safe edit boundary: a112 does not edit this function. `hybrid.Client.GetQuote` and the console depend on its exact silence-on-empty behaviour; L1c adds a sibling strict reader instead.
- High-risk impact: no by role, yes by adjacency - same shared production token path and shared quota as `Client.Orderbook`, and L1c needs both endpoints for one seal, so a quote read costs two GETs against that quota rather than one.
- Untested branch: B1 (the `c.get` error path) and the empty-body case of the fall-through. The populated success path is pinned by `TestPricesIntegration`; the package suite is green (`go test ./internal/official -count=1`, 351 tests, exit 0, 2026-08-18).
