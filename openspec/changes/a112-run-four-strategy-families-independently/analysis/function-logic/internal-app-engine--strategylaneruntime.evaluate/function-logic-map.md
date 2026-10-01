# Function Logic Map: `strategyLaneRuntime.evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`
- Signature: `strategyLaneRuntime.evaluate(params=5, results=1)`
- Source range: `199:1`–`232:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 물결 번호를 시장 인자로 record 에 넘긴다 — 공유 계수기 변이 P05 CAUGHT.

## Branches and early returns

- Exact AST return nodes: `203:3, 209:3, 227:3, 231:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 202:2 | nil 런타임 |
| B2 | if | 208:2 | 복구 실패 → 오류 |
| B3 | range | 213:2 | 이 시장 네 레인 순회 |
| B4 | range | 215:3 | 레인 입력 찾기 |
| B5 | if | 216:4 | 자기 제안 소유 |
| B6 | if | 226:2 | 잠금 기록 실패 → 오류 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.recoverMarketLanes` | 208:12 |
| `runtime.lanesFor` | 211:11 |
| `make` | 212:18 |
| `len` | 212:53 |
| `lane.Owns` | 216:7 |
| `append` | 221:18 |
| `runtime.runLane` | 221:39 |
| `runtime.record` | 223:2 |
| `runtime.persistMarketLatches` | 226:12 |
| `runtime.clk.Now` | 226:76 |
| `runtime.staleLatchError` | 231:9 |

## State mutations and fallbacks

- 편집 전과 같음(레인 · 원장 잠금).

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
