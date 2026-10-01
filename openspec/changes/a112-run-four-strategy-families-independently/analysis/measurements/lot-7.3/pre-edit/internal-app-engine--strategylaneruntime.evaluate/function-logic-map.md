# Function Logic Map (편집 전): `evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`
- Signature: `strategyLaneRuntime.evaluate(params=5, results=1)`
- Source range: `191:1`–`224:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: record 호출에 시장을 넘긴다(한 줄). 순서(복구 → 돌기 → 기록 → 잠금 남기기) 불변.

## Branches and early returns

- Exact AST return nodes: `195:3`, `201:3`, `219:3`, `223:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 194:2 | `if runtime == nil {` |
| B2 | if | 200:2 | `if err := runtime.recoverMarketLanes(ctx, market, activationGeneration); err != nil {` |
| B3 | range | 205:2 | `for _, lane := range lanes {` |
| B4 | range | 207:3 | `for _, candidate := range inputs {` |
| B5 | if | 208:4 | `if lane.Owns(candidate.Proposal) {` |
| B6 | if | 218:2 | `if err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.recoverMarketLanes` | 200:12 |
| `runtime.lanesFor` | 203:11 |
| `make` | 204:18 |
| `len` | 204:53 |
| `lane.Owns` | 208:7 |
| `append` | 213:18 |
| `runtime.runLane` | 213:39 |
| `runtime.record` | 215:2 |
| `runtime.persistMarketLatches` | 218:12 |
| `runtime.clk.Now` | 218:76 |
| `runtime.staleLatchError` | 223:9 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
