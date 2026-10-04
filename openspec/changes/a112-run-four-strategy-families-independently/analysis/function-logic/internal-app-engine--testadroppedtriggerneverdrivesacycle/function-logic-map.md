# Function Logic Map: `TestADroppedTriggerNeverDrivesACycle (시험)`

- Source: `internal/app/engine/a112_lane_runtime_test.go`
- Source SHA-256: `c0eec4047bca78202787fdabf8d0ef78eaee26237ce64fb46fd65edf901a5779`
- Signature: `TestADroppedTriggerNeverDrivesACycle(params=1, results=0)`
- Source range: `143:1`–`179:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 150:2 | 시험 자신의 갈래 |
| B2 | if | 155:2 | 시험 자신의 갈래 |
| B3 | if | 162:2 | 시험 자신의 갈래 |
| B4 | range | 165:2 | 시험 자신의 갈래 |
| B5 | if | 166:3 | 시험 자신의 갈래 |
| B6 | if | 169:3 | 시험 자신의 갈래 |
| B7 | if | 172:3 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneRuntimeFixture` | 144:19 |
| `runtime.lanesFor` | 145:10 |
| `Cadence` | 146:13 |
| `lane.Policy` | 146:13 |
| `runtime.evaluate` | 149:2 |
| `context.Background` | 149:19 |
| `lane.Dropped` | 150:12 |
| `t.Fatalf` | 151:3 |
| `runtime.evaluate` | 154:2 |
| `context.Background` | 154:19 |
| `lane.Pending` | 155:18 |
| `QueueDepth` | 155:34 |
| `lane.Policy` | 155:34 |
| `t.Fatalf` | 156:3 |
| `fake.Advance` | 159:2 |
| `runtime.evaluate` | 160:2 |
| `context.Background` | 160:19 |
| `lane.Dropped` | 162:12 |
| `t.Fatal` | 163:3 |
| `runtime.observations` | 165:30 |
| `lane.Key` | 166:25 |
| `t.Fatalf` | 170:4 |
| `t.Fatalf` | 173:4 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
