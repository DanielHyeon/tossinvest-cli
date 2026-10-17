# Function Logic Map: `adaptPrices`

- Source: `internal/official/market_reads.go`
- Source SHA-256: `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea` (current worktree; verified with `sha256sum` 2026-08-18, equal to `source_sha256` in `ast.json`)
- Signature: `adaptPrices(raw []apiPrice) []domain.Quote` (`ast.json`: `adaptPrices(params=1, results=1)`)
- Source range: `168:1`-`180:2`
- AST counts: branches 1, returns 1, calls 6, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief (the `last` half of the quote seal). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw` | `[]apiPrice{Symbol, LastPrice, Currency, Timestamp string}` | `/api/v1/prices` `result` array | empty slice ⇒ empty result, no error |
| `raw[i].LastPrice` | decimal string | the response | `parseDecimal` fail-open to `0` |
| `raw[i].Timestamp` | RFC3339 string, **nullable** per the schema comment at 133 | the response | **dropped** - never referenced in this body |

- `apiPrice` (137-142) declares `Timestamp`, and this body never reads it: the AST call list is `make`, `len`, `append`, `parseDecimal`, `time.Now`, `UTC` - no `time.Parse`. `FetchedAt` at 175 is the local process clock.
- Consequence for L1c: `last` and `bid`/`ask` come from **two different endpoints** (`/api/v1/prices` and `/api/v1/orderbook`), and on the current adapters neither endpoint's own instant survives. A quote seal built from these two reads has no measured coherence between its halves; the binding rule and the skew bound must therefore be decided from raw bytes, not from either adapter.

## Branches and early returns

Exact AST return node: `179`. One return, no early exit.

- B1 (`range` at 170:2) - per element `append` at 171:3 of `domain.Quote{Symbol, Last: parseDecimal(p.LastPrice) (173:10), Currency, FetchedAt: time.Now().UTC() (175:15)}`. Zero elements ⇒ empty slice.
- No per-element validation: symbol echo, currency and price are copied without checking that the returned rows correspond to the requested symbols, in order, or at all. A batch request for N symbols may return fewer rows and nothing here notices.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 169:9 |
| `len` | 169:33 |
| `append` | 171:9 |
| `parseDecimal` | 173:15 |
| `UTC` | 175:15 |
| `time.Now` | 175:15 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseDecimal` (173:10) | `lastPrice` string -> `float64` | fail-open `0` (`internal-official--parsedecimal` bundle) | `ast.json` call |
| `make`/`len`/`append` | build the result slice | none | `ast.json` calls at 169, 171 |
| `time.Now().UTC()` (175:15) | stamp `FetchedAt` | none; process clock, no authority | `ast.json` call |

## State mutations and fallbacks

- No package state, no I/O. The fallbacks are inherited from `parseDecimal` (silent `0`) and from the dropped `Timestamp` (no broker instant).
- `Name`, `MarketCode`, `Market`, `Change`, `Volume` are left zero by construction (comment at 177).

## Safety conclusion

- Safe edit boundary: a112 does not edit this function. It backs `hybrid.Client.GetQuote` and the console quote screen; L1c adds its own reader rather than tightening a display path.
- High-risk impact: no by role (no order, sizing or protection decision). Epistemic risk for a112: it drops the response `timestamp` and echoes rows without matching them to the requested symbols, so nothing downstream can tell a stale or mismatched price from a fresh one. A quote seal built on that silence would be unfalsifiable.
- Untested branch: none of B1's paths are untested in the adapter itself (`TestAdaptPricesUnit`, `TestAdaptPricesEmpty`); what is untested is the absence itself - no test asserts the requested-vs-returned symbol correspondence, because the code does not implement it. Package suite green (`go test ./internal/official -count=1`, 351 tests, exit 0, 2026-08-18).
