# Function Logic Map: `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts(params=1, results=0)`
- Source range: `108:1`–`132:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 113:2 | 시험 자신의 갈래 |
| B2 | if | 121:2 | 시험 자신의 갈래 |
| B3 | if | 129:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 109:13 |
| `context.Background` | 110:9 |
| `mustProductionStrategyLanes` | 112:13 |
| `t.Fatal` | 114:3 |
| `runtime.lanesFor` | 116:10 |
| `lane.Key` | 117:9 |
| `latchOneLane` | 118:2 |
| `runtime.evaluate` | 121:12 |
| `t.Fatalf` | 122:3 |
| `mustProductionStrategyLanes` | 128:15 |
| `Latched` | 129:6 |
| `laneByKey` | 129:6 |
| `t.Fatal` | 130:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
