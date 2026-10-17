# Function Logic Map: `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith (시험)`

- Source: `internal/app/engine/a112_lane_coordinator_projection_test.go`
- Source SHA-256: `a1f16eb741b42836ad3606bcb3fdf6d1725ae56579ef6bbb03b9b43dc6ef42ee`
- Signature: `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith(params=1, results=0)`
- Source range: `124:1`–`150:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 127:2 | 시험 자신의 갈래 |
| B2 | range | 130:2 | 시험 자신의 갈래 |
| B3 | if | 132:3 | 시험 자신의 갈래 |
| B4 | if | 135:3 | 시험 자신의 갈래 |
| B5 | if | 140:3 | 시험 자신의 갈래 |
| B6 | else | 146:10 | 시험 자신의 갈래 |
| B7 | if | 141:4 | 시험 자신의 갈래 |
| B8 | if | 146:10 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `a112LaneProjectionContext` | 125:17 |
| `strategyrouter.FamilyActivationForTest` | 126:16 |
| `strategyrouter.AllFourFamiliesForTest` | 126:83 |
| `lanes.evaluate` | 127:12 |
| `context.Background` | 127:27 |
| `t.Fatal` | 128:3 |
| `a112Read` | 130:23 |
| `t.Fatalf` | 136:4 |
| `string` | 142:22 |
| `t.Fatalf` | 143:5 |
| `t.Fatalf` | 147:4 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
