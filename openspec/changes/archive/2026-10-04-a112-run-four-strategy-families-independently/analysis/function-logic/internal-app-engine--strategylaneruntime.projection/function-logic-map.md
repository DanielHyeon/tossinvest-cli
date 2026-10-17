# Function Logic Map: `projection`

- Source: `internal/app/engine/strategy_lane_projection.go`
- Source SHA-256: `713fd68268cb36340d9c990c2d213c56a3b9d72c1342f83e2b845f09254978ef`
- Signature: `strategyLaneRuntime.projection(params=0, results=1)`
- Source range: `32:1`–`53:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 함수 하나 — 경계 핀 (e)(f) · 나이 등식 · 의도된 간극.

## Branches and early returns

- Exact AST return nodes: `34:3, 52:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 33:2 | 런타임 nil |
| B2 | range | 42:2 | 레인 순회 |
| B3 | if | 47:3 | **(새)** 쓸 수 있는 shadow 관측 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.clk.Now` | 38:9 |
| `runtime.mu.RLock` | 39:2 |
| `runtime.mu.RUnlock` | 40:8 |
| `make` | 41:12 |
| `len` | 41:64 |
| `lane.Key` | 43:10 |
| `StrategyMarket` | 46:13 |
| `shadowObservationUsable` | 47:62 |
| `append` | 50:12 |
| `strategyLaneProjection` | 50:27 |

## State mutations and fallbacks

- 없음(읽기).

## Safety conclusion

- 읽기 전용 — 레인 · 관측을 바꾸지 않는다.
