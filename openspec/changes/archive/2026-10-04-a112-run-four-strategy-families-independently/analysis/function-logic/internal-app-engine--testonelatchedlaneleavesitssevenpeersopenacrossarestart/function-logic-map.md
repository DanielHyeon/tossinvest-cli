# Function Logic Map: `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart(params=1, results=0)`
- Source range: `206:1`–`233:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 213:2 | 시험 자신의 갈래 |
| B2 | if | 214:3 | 시험 자신의 갈래 |
| B3 | range | 222:2 | 시험 자신의 갈래 |
| B4 | if | 223:3 | 시험 자신의 갈래 |
| B5 | if | 225:4 | 시험 자신의 갈래 |
| B6 | if | 230:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 207:13 |
| `context.Background` | 208:9 |
| `mustProductionStrategyLanes` | 209:13 |
| `runtime.lanesFor` | 210:10 |
| `lane.Key` | 211:9 |
| `latchOneLane` | 212:2 |
| `runtime.evaluate` | 214:13 |
| `t.Fatalf` | 215:4 |
| `mustProductionStrategyLanes` | 220:15 |
| `peer.Latched` | 223:6 |
| `peer.Key` | 225:7 |
| `t.Fatalf` | 226:5 |
| `Parts` | 226:72 |
| `peer.Key` | 226:72 |
| `t.Fatalf` | 231:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
