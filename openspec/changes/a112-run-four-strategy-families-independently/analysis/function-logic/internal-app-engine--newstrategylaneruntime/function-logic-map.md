# Function Logic Map: `newStrategyLaneRuntime`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `333970fa5e15db9741cc10da836922c79763f4b1b87bf62ddbc1af2fba9c6462`
- Signature: `newStrategyLaneRuntime(params=3, results=1)`
- Source range: `105:1`–`120:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 맵 초기화만 더함.

## Branches and early returns

- Exact AST return nodes: `107:3, 112:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 106:2 | 시계 nil → nil 런타임 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyworker.ProductionLanes` | 109:11 |
| `make` | 113:19 |
| `len` | 113:72 |
| `make` | 115:19 |
| `make` | 116:19 |
| `make` | 117:19 |
| `make` | 118:19 |
| `make` | 119:19 |

## State mutations and fallbacks

- 새 런타임 값.

## Safety conclusion

- 레인 목록 · 원장 · 계좌 무변경.
