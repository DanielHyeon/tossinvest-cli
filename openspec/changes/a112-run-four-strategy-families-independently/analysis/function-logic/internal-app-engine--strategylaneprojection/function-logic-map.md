# Function Logic Map: `strategyLaneProjection`

- Source: `internal/app/engine/strategy_lane_projection.go`
- Source SHA-256: `04dcd7ed4366c18c5ac8b8c0cc490b944f5287dee4db481c35cfec07e6173d70`
- Signature: `strategyLaneProjection(params=3, results=1)`
- Source range: `45:1`–`87:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.5).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 한 행은 한 잠금으로 읽힌다 — AST 핀 `TestTheLaneProjectionReadsALaneRowUnderOneLock`(Status 1 회 · 상태 접근자 0, 변이 R05 CAUGHT); 행 불변식 동시성 시험 `TestAProjectedLaneRowIsNeverTorn`(-race, 비결정이라 결정적 핀은 AST — Manager 승인).

## Branches and early returns

- Exact AST return nodes: `62:3, 86:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 61:2 | 미관측 레인 → 관측 사실 없이 반환 |
| B2 | if | 68:2 | 투입이 들어간 물결 → 시작 |
| B3 | if | 71:3 | 연 사이클 → 결과 · 비정상 |
| B4 | if | 73:4 | 결과 있음 |
| B5 | if | 79:4 | REFUSED 결과만 거절 코드 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 49:22 |
| `lane.Policy` | 49:34 |
| `lane.Status` | 49:49 |
| `strategyprojection.LaneHealth` | 50:12 |
| `policy.Version` | 51:23 |
| `Milliseconds` | 51:41 |
| `policy.CycleDeadline` | 51:41 |
| `strategyprojection.Market` | 52:60 |
| `string` | 52:107 |
| `string` | 53:62 |
| `lane.Horizon` | 53:69 |
| `strategyprojection.LaneRuntime` | 55:12 |
| `lane.Runtime` | 55:43 |
| `strategyprojection.NormalizedText` | 57:17 |
| `projectionTime` | 60:14 |
| `projectionTime` | 60:61 |
| `strategyprojection.State` | 65:35 |
| `strategyprojection.State` | 65:82 |
| `strategyprojection.LaneTrigger` | 66:13 |
| `strategyprojection.LaneStart` | 69:12 |
| `strategyprojection.LaneOutcome` | 74:16 |
| `projectionOptional` | 80:21 |
| `string` | 80:40 |
| `projectionOptional` | 84:25 |
| `projectionOptional` | 85:25 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 전용.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영.
