# Function Logic Map: `TestEveryLaneStaysDormantOnAProposalItActuallyOwns (시험)`

- Source: `internal/app/engine/a112_lane_dependency_test.go`
- Source SHA-256: `48c82734876cae38abe05263bdd539c3ab55df16667c99641ae7e8b378e07b8e`
- Signature: `TestEveryLaneStaysDormantOnAProposalItActuallyOwns(params=1, results=0)`
- Source range: `79:1`–`99:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 83:2 | 시험 자신의 갈래 |
| B2 | if | 84:3 | 시험 자신의 갈래 |
| B3 | if | 88:3 | 시험 자신의 갈래 |
| B4 | if | 92:3 | 시험 자신의 갈래 |
| B5 | if | 96:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneRuntimeForDependency` | 80:13 |
| `runtime.evaluate` | 81:6 |
| `context.Background` | 81:23 |
| `laneOwnedInputs` | 81:101 |
| `runtime.observations` | 83:30 |
| `t.Fatalf` | 89:4 |
| `t.Fatalf` | 93:4 |
| `t.Fatalf` | 97:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
