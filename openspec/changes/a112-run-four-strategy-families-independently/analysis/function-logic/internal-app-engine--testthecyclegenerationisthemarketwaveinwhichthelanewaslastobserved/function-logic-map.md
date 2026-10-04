# Function Logic Map: `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved (시험)`

- Source: `internal/app/engine/a112_lane_coordinator_projection_test.go`
- Source SHA-256: `a1f16eb741b42836ad3606bcb3fdf6d1725ae56579ef6bbb03b9b43dc6ef42ee`
- Signature: `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved(params=1, results=0)`
- Source range: `80:1`–`121:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 84:2 | 시험 자신의 갈래 |
| B2 | range | 88:2 | 시험 자신의 갈래 |
| B3 | if | 90:3 | 시험 자신의 갈래 |
| B4 | if | 93:3 | 시험 자신의 갈래 |
| B5 | if | 99:2 | 시험 자신의 갈래 |
| B6 | if | 102:2 | 시험 자신의 갈래 |
| B7 | range | 105:2 | 시험 자신의 갈래 |
| B8 | if | 107:3 | 시험 자신의 갈래 |
| B9 | if | 110:3 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `a112LaneProjectionContext` | 81:20 |
| `strategyworker.ProductionRuntimePolicy` | 82:12 |
| `context.Background` | 83:9 |
| `lanes.evaluate` | 84:12 |
| `t.Fatal` | 85:3 |
| `a112Read` | 87:12 |
| `uint64` | 89:11 |
| `t.Fatalf` | 94:4 |
| `fake.Advance` | 98:2 |
| `policy.Cadence` | 98:15 |
| `lanes.evaluate` | 99:12 |
| `t.Fatal` | 100:3 |
| `lanes.evaluate` | 102:12 |
| `t.Fatal` | 103:3 |
| `a112Read` | 105:23 |
| `uint64` | 106:11 |
| `policy.Version` | 115:56 |
| `Milliseconds` | 116:60 |
| `policy.CycleDeadline` | 116:60 |
| `t.Fatalf` | 118:4 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
