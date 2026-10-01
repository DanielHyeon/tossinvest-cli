# Function Logic Map (편집 전): `runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- Source range: `260:1`–`287:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.5, Manager 판정 D1=(A) · D2 · C1~C3 2026-10-01).

## Inputs and invariants

- 편집 계획: step 을 런타임 seam(runtime.laneStep — nil 이면 strategyFamilyLaneStep)에서 얻는다. 레인 호출 순서 · 횟수 불변.

## Branches and early returns

- Exact AST return nodes: `270:3`, `286:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 268:2 | `if observation.Trigger != strategyworker.TriggerEnqueued {` |
| B2 | if | 281:2 | `if bounded.Err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 265:46 |
| `lane.Desired` | 265:67 |
| `lane.Effective` | 265:103 |
| `strategyLaneEvidenceDigest` | 266:57 |
| `lane.Offer` | 267:24 |
| `lane.Health` | 269:24 |
| `lane.RunBounded` | 272:20 |
| `strategyFamilyLaneStep` | 272:48 |
| `bounded.Err.Error` | 282:25 |
| `lane.Health` | 284:23 |

## Safety conclusion

- High-risk 아님 — 생산 기본값 불변(seam 은 시험만 채움).
