# Function Logic Map: `parseDecimal`

- Source: `internal/official/market_reads.go`
- Source SHA-256: `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea` (current worktree; verified with `sha256sum` 2026-08-18, equal to `source_sha256` in `ast.json`)
- Signature: `parseDecimal(s string) float64` (`ast.json`: `parseDecimal(params=1, results=1)`)
- Source range: `15:1`-`24:2`
- AST counts: branches 2, returns 3, calls 1, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief (official quote producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | any string; in practice a broker decimal string (`"72100"`, `"185.70"`) | the official JSON response field, already decoded into a Go `string` by `encoding/json` | never returns an error; every failure collapses to `0` |

- The only call is `strconv.ParseFloat(s, 64)` at 19:12. There is no error channel out of this function: the signature has one result and it is a `float64`.
- Consequence relied on by the L1c brief: a value this function returns cannot be distinguished from a value the broker actually sent as `0`, and a value that round-trips through `float64` is no longer the bytes the broker sent. Both properties are structural, not incidental.

## Branches and early returns

Exact AST return nodes: `17, 21, 23`.

- B1 (`if` at 16:2) - empty string. Returns `0` at 17:3 without calling the parser. An absent JSON string field decodes to `""` and therefore also lands here.
- B2 (`if` at 20:2) - `strconv.ParseFloat` returned an error (non-numeric, over/underflow, `NaN` text, leading/trailing space). Returns `0` at 21:3. The error value is discarded; nothing is logged and no caller can observe it.
- Fall-through - returns the parsed `float64` at 23:2. IEEE-754 binary64 cannot represent every decimal string exactly, so the returned value is a rounding of the received bytes, and the received bytes are not retained by this function.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strconv.ParseFloat` | 19:12 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strconv.ParseFloat` | convert the decimal string to `float64` | error is swallowed at B2 and mapped to `0`; no retry | `ast.json` call at 19:12 |

## State mutations and fallbacks

- No package state, no I/O, no clock. Pure function.
- The fallback IS the defect surface for evidence work: `0` is both a legitimate broker value and this function's universal error code. `internal/strategyevidence`'s `minorFromRawDecimal` exists precisely because evidence may not be minted through a fail-open float conversion (a112 decision 9).

## Safety conclusion

- Safe edit boundary: a112 does not edit this function. It is called by every official adapter that receives a numeric string, so any body change would move the display semantics of the whole read surface at once; that is out of L1c's scope and out of its ownership.
- High-risk impact: no by role (display/domain projection only, no order, sizing or protection decision), but the fail-open `0` at B2 is exactly the shape evidence work must not inherit. L1c reads decimals through `strategyevidence.minorFromRawDecimal` (integer arithmetic, typed refusal), never through this function.
- Untested branch: B2 (parse error). B1 is only reached incidentally and the fall-through is pinned by `TestAdaptPricesUnit`/`TestAdaptOrderbookUnit`; the package suite is green (`go test ./internal/official -count=1`, 351 tests, exit 0, 2026-08-18).
