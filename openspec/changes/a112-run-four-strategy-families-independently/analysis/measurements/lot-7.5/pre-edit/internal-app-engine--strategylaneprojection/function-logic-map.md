# Function Logic Map (편집 전): `strategyLaneProjection`

- Source: `internal/app/engine/strategy_lane_projection.go`
- Source SHA-256: `b51783c00c67b9a1072683006251c6d1f7a50d2f18ec97f94acb77eb13f1c1a1`
- Signature: `strategyLaneProjection(params=3, results=1)`
- Source range: `45:1`–`85:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.5, Manager 판정 D1=(A) · D2 · C1~C3 2026-10-01).

## Inputs and invariants

- 편집 계획: 레인 행을 접근자별 잠금 대신 Lane.Status() 한 잠금으로 읽는다(D2 — 찢긴 행 제거). 정책 · horizon · runtime 은 불변 값이라 그대로.

## Branches and early returns

- Exact AST return nodes: `60:3`, `84:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 59:2 | `if !observed {` |
| B2 | if | 66:2 | `if observation.Trigger == strategyworker.TriggerEnqueued {` |
| B3 | if | 69:3 | `if observation.Start == strategyworker.StartAdmitted {` |
| B4 | if | 71:4 | `if observation.Outcome != "" {` |
| B5 | if | 77:4 | `if observation.Outcome == strategyworker.OutcomeRefused {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 47:17 |
| `lane.Policy` | 47:29 |
| `strategyprojection.LaneHealth` | 48:12 |
| `lane.Health` | 48:42 |
| `policy.Version` | 49:23 |
| `Milliseconds` | 49:41 |
| `policy.CycleDeadline` | 49:41 |
| `strategyprojection.Market` | 50:60 |
| `string` | 50:107 |
| `string` | 51:62 |
| `lane.Horizon` | 51:69 |
| `strategyprojection.LaneRuntime` | 53:12 |
| `lane.Runtime` | 53:43 |
| `lane.ConsecutiveFailures` | 54:24 |
| `lane.LatchRevision` | 54:67 |
| `strategyprojection.NormalizedText` | 55:17 |
| `lane.FirstFailure` | 55:51 |
| `lane.Pending` | 56:17 |
| `lane.Dropped` | 56:42 |
| `lane.Abandoned` | 56:69 |
| `projectionTime` | 58:14 |
| `lane.NextDue` | 58:29 |
| `projectionTime` | 58:64 |
| `lane.RestartNotBefore` | 58:79 |
| `strategyprojection.State` | 63:35 |
| `strategyprojection.State` | 63:82 |
| `strategyprojection.LaneTrigger` | 64:13 |
| `strategyprojection.LaneStart` | 67:12 |
| `strategyprojection.LaneOutcome` | 72:16 |
| `projectionOptional` | 78:21 |
| `string` | 78:40 |
| `projectionOptional` | 82:25 |
| `projectionOptional` | 83:25 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영.
