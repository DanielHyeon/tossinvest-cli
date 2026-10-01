# Function Logic Map (편집 전): `evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`
- Signature: `strategyLaneRuntime.evaluate(params=5, results=1)`
- Source range: `199:1`–`232:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.5, Manager 판정 D1=(A) · D2 · C1~C3 2026-10-01).

## Inputs and invariants

- 편집 계획: 레인 넷을 순차 대신 동시 실행 + join(시장 주기 안). 관측은 레인 순서 그대로 색인으로 모음. 레인 goroutine 의 panic 은 join 뒤 시장 주기 goroutine 에서 다시 던져 기존 회복 경로를 탐. 순서(복구 → 돌기 → 기록 → 잠금 남기기) 불변.

## Branches and early returns

- Exact AST return nodes: `203:3`, `209:3`, `227:3`, `231:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 202:2 | `if runtime == nil {` |
| B2 | if | 208:2 | `if err := runtime.recoverMarketLanes(ctx, market, activationGeneration); err != nil {` |
| B3 | range | 213:2 | `for _, lane := range lanes {` |
| B4 | range | 215:3 | `for _, candidate := range inputs {` |
| B5 | if | 216:4 | `if lane.Owns(candidate.Proposal) {` |
| B6 | if | 226:2 | `if err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {` |

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

## Safety conclusion

- High-risk 인접(레인 런타임 동시성) — 주문 · 원장 쓰기 없음. 레인은 서로 상태를 공유하지 않는다(Lane 주석). 동시 실행은 같은 레인을 두 번 돌리지 않는다(레인당 goroutine 하나).
