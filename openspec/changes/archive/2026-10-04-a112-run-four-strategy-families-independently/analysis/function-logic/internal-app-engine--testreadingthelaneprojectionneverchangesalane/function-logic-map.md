# Function Logic Map: `TestReadingTheLaneProjectionNeverChangesALane (시험)`

- Source: `internal/app/engine/a112_lane_coordinator_projection_test.go`
- Source SHA-256: `a1f16eb741b42836ad3606bcb3fdf6d1725ae56579ef6bbb03b9b43dc6ef42ee`
- Signature: `TestReadingTheLaneProjectionNeverChangesALane(params=1, results=0)`
- Source range: `209:1`–`237:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 211:2 | 시험 자신의 갈래 |
| B2 | if | 217:2 | 시험 자신의 갈래 |
| B3 | range | 221:2 | 시험 자신의 갈래 |
| B4 | if | 225:2 | 시험 자신의 갈래 |
| B5 | if | 228:2 | 시험 자신의 갈래 |
| B6 | if | 231:2 | 시험 자신의 갈래 |
| B7 | if | 234:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `a112LaneProjectionContext` | 210:20 |
| `lanes.evaluate` | 211:12 |
| `context.Background` | 211:27 |
| `t.Fatal` | 212:3 |
| `fake.Advance` | 214:2 |
| `a112LaneStates` | 215:26 |
| `lanes.observations` | 215:49 |
| `c.Journal.OpenStrategyLaneLatches` | 216:15 |
| `context.Background` | 216:49 |
| `t.Fatal` | 218:3 |
| `a112Read` | 220:11 |
| `a112Read` | 222:3 |
| `a112Read` | 224:10 |
| `a112LaneStates` | 225:12 |
| `reflect.DeepEqual` | 225:36 |
| `t.Fatalf` | 226:3 |
| `lanes.observations` | 228:12 |
| `reflect.DeepEqual` | 228:35 |
| `t.Fatalf` | 229:3 |
| `c.Journal.OpenStrategyLaneLatches` | 231:19 |
| `context.Background` | 231:53 |
| `len` | 231:110 |
| `len` | 231:124 |
| `t.Fatalf` | 232:3 |
| `len` | 232:83 |
| `len` | 232:94 |
| `reflect.DeepEqual` | 234:6 |
| `t.Fatal` | 235:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
