# Function Logic Map: `strategyLaneProjection`

- Source: `internal/app/engine/strategy_lane_projection.go`
- Source SHA-256: `713fd68268cb36340d9c990c2d213c56a3b9d72c1342f83e2b845f09254978ef`
- Signature: `strategyLaneProjection(params=4, results=1)`
- Source range: `57:1`–`104:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- ON 관측은 SHADOW 가 아니다 — `TestNeitherTheStepNorTheProjectionShadowsAnOnLane`(변이 S30).

## Branches and early returns

- Exact AST return nodes: `75:3, 103:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 74:2 | 미관측 레인 |
| B2 | if | 81:2 | 투입이 들어간 물결 |
| B3 | if | 84:3 | 연 사이클 |
| B4 | if | 86:4 | 결과 있음 |
| B5 | if | 92:4 | REFUSED 결과만 거절 코드 |
| B6 | if | 99:2 | **(새)** SHADOW — 관측된 OFF/OFF 레인만 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 62:22 |
| `lane.Policy` | 62:34 |
| `lane.Status` | 62:49 |
| `strategyprojection.LaneHealth` | 63:12 |
| `policy.Version` | 64:23 |
| `Milliseconds` | 64:41 |
| `policy.CycleDeadline` | 64:41 |
| `strategyprojection.Market` | 65:60 |
| `string` | 65:107 |
| `string` | 66:62 |
| `lane.Horizon` | 66:69 |
| `strategyprojection.LaneRuntime` | 68:12 |
| `lane.Runtime` | 68:43 |
| `strategyprojection.NormalizedText` | 70:17 |
| `projectionTime` | 73:14 |
| `projectionTime` | 73:61 |
| `strategyprojection.State` | 78:35 |
| `strategyprojection.State` | 78:82 |
| `strategyprojection.LaneTrigger` | 79:13 |
| `strategyprojection.LaneStart` | 82:12 |
| `strategyprojection.LaneOutcome` | 87:16 |
| `projectionOptional` | 93:21 |
| `string` | 93:40 |
| `projectionOptional` | 97:25 |
| `projectionOptional` | 98:25 |
| `strategyprojection.LaneShadowOutcome` | 100:14 |

## State mutations and fallbacks

- 없음(값 구성).

## Safety conclusion

- 읽기 전용 — SHADOW 는 승격하지 않는다(desired/effective 는 관측 값 그대로).
