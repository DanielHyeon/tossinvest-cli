# Function Logic Map (편집 전): `newStrategyLaneRuntime`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`
- Signature: `newStrategyLaneRuntime(params=3, results=1)`
- Source range: `95:1`–`103:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 v3.3 §5 N5: 생성자에서 shadowEpoch 맵(과 시장 칸 맵) 초기화 — FLM 확정 목록 밖에서 새로 잡힌 기존 함수(nil 맵 대입 panic 차단).

## Branches and early returns

- Exact AST return nodes: `97:3`, `100:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 96:2 | `if clk == nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyworker.ProductionLanes` | 99:11 |
| `make` | 101:13 |
| `len` | 101:66 |

## Safety conclusion

- 레인 런타임 생성 — 레인 목록 · 원장 · 계좌 무변경, 맵 초기화 추가뿐.
