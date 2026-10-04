# Function Logic Map (편집 전): `strategyLaneProjection`

- Source: `internal/app/engine/strategy_lane_projection.go`
- Source SHA-256: `04dcd7ed4366c18c5ac8b8c0cc490b944f5287dee4db481c35cfec07e6173d70`
- Signature: `strategyLaneProjection(params=3, results=1)`
- Source range: `45:1`–`87:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §7: 사용 가능한 shadow 관측 ∧ ShadowEligible 이면 Runtime=SHADOW · shadowOutcome, 아니면 UNOBSERVED · null.

## Branches and early returns

- Exact AST return nodes: `62:3`, `86:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 61:2 | `if !observed {` |
| B2 | if | 68:2 | `if observation.Trigger == strategyworker.TriggerEnqueued {` |
| B3 | if | 71:3 | `if observation.Start == strategyworker.StartAdmitted {` |
| B4 | if | 73:4 | `if observation.Outcome != "" {` |
| B5 | if | 79:4 | `if observation.Outcome == strategyworker.OutcomeRefused {` |

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

## Safety conclusion

- 읽기 전용 투영 — 기존 필드 무변경, SHADOW 는 projection 어휘에만.
