# Function Logic Map: `evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `333970fa5e15db9741cc10da836922c79763f4b1b87bf62ddbc1af2fba9c6462`
- Signature: `strategyLaneRuntime.evaluate(params=6, results=1)`
- Source range: `218:1`–`277:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 묶음 사용은 record 인자 한 자리 — AST 핀.

## Branches and early returns

- Exact AST return nodes: `222:3, 228:3, 272:3, 276:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 221:2 | 런타임 nil |
| B2 | if | 227:2 | 복구 오류 |
| B3 | range | 247:2 | 레인 순회(동시) |
| B4 | range | 249:3 | 입력 찾기 |
| B5 | if | 250:4 | 레인 소유 |
| B6 | range | 263:2 | panic 수거 순회 |
| B7 | if | 264:3 | panic 재던짐 |
| B8 | if | 271:2 | latch 저장 오류 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.recoverMarketLanes` | 227:12 |
| `runtime.lanesFor` | 230:11 |
| `make` | 244:18 |
| `len` | 244:50 |
| `make` | 245:12 |
| `len` | 245:24 |
| `lane.Owns` | 250:7 |
| `join.Add` | 255:3 |
| `(unnamed)` | 256:6 |
| `join.Done` | 257:10 |
| `(unnamed)` | 258:10 |
| `recover` | 258:35 |
| `runtime.runLane` | 259:26 |
| `join.Wait` | 262:2 |
| `panic` | 265:4 |
| `runtime.record` | 268:2 |
| `runtime.persistMarketLatches` | 271:12 |
| `runtime.clk.Now` | 271:76 |
| `runtime.staleLatchError` | 276:9 |

## State mutations and fallbacks

- record 위임.

## Safety conclusion

- 레인 실행 · 관측 · latch 저장 불변.
