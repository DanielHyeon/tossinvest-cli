# Function Logic Map: `TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed(params=1, results=0)`
- Source range: `246:1`–`277:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 253:2 | 시험 자신의 갈래 |
| B2 | if | 259:2 | 시험 자신의 갈래 |
| B3 | if | 263:2 | 시험 자신의 갈래 |
| B4 | if | 267:2 | 시험 자신의 갈래 |
| B5 | if | 271:2 | 시험 자신의 갈래 |
| B6 | if | 274:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 247:13 |
| `context.Background` | 248:9 |
| `fake.Now` | 252:83 |
| `c.Journal.RecordStrategyLaneLatch` | 253:15 |
| `t.Fatalf` | 254:3 |
| `mustProductionStrategyLanes` | 256:13 |
| `runtime.evaluate` | 259:12 |
| `t.Fatal` | 260:3 |
| `runtime.evaluate` | 263:12 |
| `t.Fatalf` | 264:3 |
| `runtime.evaluate` | 267:12 |
| `t.Fatalf` | 268:3 |
| `c.Journal.OpenStrategyLaneLatches` | 270:15 |
| `t.Fatalf` | 272:3 |
| `len` | 274:5 |
| `t.Fatalf` | 275:3 |
| `len` | 275:72 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
