# Function Logic Map: `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart(params=1, results=0)`
- Source range: `180:1`–`202:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 187:2 | 시험 자신의 갈래 |
| B2 | if | 193:2 | 시험 자신의 갈래 |
| B3 | if | 199:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 181:13 |
| `context.Background` | 182:9 |
| `mustProductionStrategyLanes` | 183:13 |
| `runtime.lanesFor` | 184:10 |
| `lane.Key` | 185:9 |
| `latchOneLane` | 186:2 |
| `runtime.evaluate` | 187:12 |
| `t.Fatalf` | 188:3 |
| `lane.Fail` | 192:2 |
| `runtime.evaluate` | 193:12 |
| `t.Fatalf` | 194:3 |
| `mustProductionStrategyLanes` | 198:15 |
| `FirstFailure` | 199:12 |
| `laneByKey` | 199:12 |
| `t.Fatalf` | 200:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
