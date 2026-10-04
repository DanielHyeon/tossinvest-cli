# Function Logic Map (편집 전): `evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`
- Signature: `strategyLaneRuntime.evaluate(params=5, results=1)`
- Source range: `199:1`–`258:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §5 ④: 인자 index 5 로 shadow 묶음을 받아 record 에 전달만(메서드 호출 · 순회 0); 복구 → 병렬 레인 → panic 재던짐 → record → latch 저장 순서 무변경.

## Branches and early returns

- Exact AST return nodes: `203:3`, `209:3`, `253:3`, `257:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 202:2 | `if runtime == nil {` |
| B2 | if | 208:2 | `if err := runtime.recoverMarketLanes(ctx, market, activationGeneration); err != nil {` |
| B3 | range | 228:2 | `for index, lane := range lanes {` |
| B4 | range | 230:3 | `for _, candidate := range inputs {` |
| B5 | if | 231:4 | `if lane.Owns(candidate.Proposal) {` |
| B6 | range | 244:2 | `for _, recovered := range panics {` |
| B7 | if | 245:3 | `if recovered != nil {` |
| B8 | if | 252:2 | `if err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.recoverMarketLanes` | 208:12 |
| `runtime.lanesFor` | 211:11 |
| `make` | 225:18 |
| `len` | 225:50 |
| `make` | 226:12 |
| `len` | 226:24 |
| `lane.Owns` | 231:7 |
| `join.Add` | 236:3 |
| `(unnamed)` | 237:6 |
| `join.Done` | 238:10 |
| `(unnamed)` | 239:10 |
| `recover` | 239:35 |
| `runtime.runLane` | 240:26 |
| `join.Wait` | 243:2 |
| `panic` | 246:4 |
| `runtime.record` | 249:2 |
| `runtime.persistMarketLatches` | 252:12 |
| `runtime.clk.Now` | 252:76 |
| `runtime.staleLatchError` | 257:9 |

## Safety conclusion

- High-risk 인접(레인 런타임) — 레인 실행 · 관측 · latch 저장 · stale 오류 무변경.
