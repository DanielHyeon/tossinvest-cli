# Function Logic Map: `TestALatchOnlyReopensForAStrictlyNewerVerifiedActivation (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestALatchOnlyReopensForAStrictlyNewerVerifiedActivation(params=1, results=0)`
- Source range: `140:1`–`175:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 147:2 | 시험 자신의 갈래 |
| B2 | for | 152:2 | 시험 자신의 갈래 |
| B3 | if | 153:3 | 시험 자신의 갈래 |
| B4 | if | 156:3 | 시험 자신의 갈래 |
| B5 | if | 162:2 | 시험 자신의 갈래 |
| B6 | if | 165:2 | 시험 자신의 갈래 |
| B7 | if | 172:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 141:13 |
| `context.Background` | 142:9 |
| `mustProductionStrategyLanes` | 143:13 |
| `runtime.lanesFor` | 144:10 |
| `lane.Key` | 145:9 |
| `latchOneLane` | 146:2 |
| `runtime.evaluate` | 147:12 |
| `t.Fatalf` | 148:3 |
| `runtime.evaluate` | 153:13 |
| `t.Fatalf` | 154:4 |
| `Latched` | 156:7 |
| `laneByKey` | 156:7 |
| `t.Fatal` | 157:4 |
| `runtime.evaluate` | 162:12 |
| `t.Fatalf` | 163:3 |
| `Latched` | 165:5 |
| `laneByKey` | 165:5 |
| `t.Fatal` | 166:3 |
| `mustProductionStrategyLanes` | 171:15 |
| `Latched` | 172:5 |
| `laneByKey` | 172:5 |
| `t.Fatal` | 173:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
