# Function Logic Map: `TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt (시험)`

- Source: `internal/app/engine/a112_lane_runtime_test.go`
- Source SHA-256: `c0eec4047bca78202787fdabf8d0ef78eaee26237ce64fb46fd65edf901a5779`
- Signature: `TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt(params=1, results=0)`
- Source range: `99:1`–`131:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 101:2 | 시험 자신의 갈래 |
| B2 | if | 105:2 | 시험 자신의 갈래 |
| B3 | range | 109:2 | 시험 자신의 갈래 |
| B4 | if | 110:3 | 시험 자신의 갈래 |
| B5 | if | 114:3 | 시험 자신의 갈래 |
| B6 | if | 117:3 | 시험 자신의 갈래 |
| B7 | if | 124:3 | 시험 자신의 갈래 |
| B8 | if | 127:3 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneRuntimeFixture` | 100:16 |
| `runtime.evaluate` | 102:3 |
| `context.Background` | 102:20 |
| `runtime.observations` | 104:18 |
| `len` | 105:5 |
| `len` | 105:26 |
| `t.Fatalf` | 106:3 |
| `len` | 107:4 |
| `len` | 107:23 |
| `t.Fatalf` | 111:4 |
| `t.Fatalf` | 115:4 |
| `t.Fatalf` | 118:4 |
| `t.Fatalf` | 125:4 |
| `t.Fatalf` | 128:4 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
