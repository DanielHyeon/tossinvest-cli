# Function Logic Map: `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther(params=1, results=0)`
- Source range: `395:1`–`435:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `418:6`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 401:2 | 시험 자신의 갈래 |
| B2 | range | 408:2 | 시험 자신의 갈래 |
| B3 | for | 413:4 | 시험 자신의 갈래 |
| B4 | if | 416:5 | 시험 자신의 갈래 |
| B5 | range | 426:2 | 시험 자신의 갈래 |
| B6 | if | 429:2 | 시험 자신의 갈래 |
| B7 | if | 432:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 396:13 |
| `context.Background` | 397:9 |
| `mustProductionStrategyLanes` | 398:13 |
| `runtime.lanesFor` | 399:8 |
| `latchOneLane` | 400:2 |
| `runtime.evaluate` | 401:12 |
| `t.Fatalf` | 402:3 |
| `make` | 405:11 |
| `make` | 406:14 |
| `wg.Add` | 409:3 |
| `(unnamed)` | 410:6 |
| `wg.Done` | 411:10 |
| `runtime.evaluate` | 416:15 |
| `close` | 423:2 |
| `wg.Wait` | 424:2 |
| `close` | 425:2 |
| `t.Fatalf` | 427:3 |
| `Latched` | 429:5 |
| `laneByKey` | 429:5 |
| `kr.Key` | 429:27 |
| `t.Fatal` | 430:3 |
| `len` | 432:12 |
| `t.Fatalf` | 433:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
