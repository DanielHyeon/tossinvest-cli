# Function Logic Map (편집 전): `runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- Source range: `286:1`–`316:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §4 FLM 목록 항목 — 예상 편집 0(관측 필드에 shadow 없음; 활성화는 record 의 칸이 들고 감). GREEN 에서 편집이 생기면 이 번들 기준으로 갱신.

## Branches and early returns

- Exact AST return nodes: `296:3`, `315:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 294:2 | `if observation.Trigger != strategyworker.TriggerEnqueued {` |
| B2 | if | 310:2 | `if bounded.Err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 291:46 |
| `lane.Desired` | 291:67 |
| `lane.Effective` | 291:103 |
| `strategyLaneEvidenceDigest` | 292:57 |
| `lane.Offer` | 293:24 |
| `lane.Health` | 295:24 |
| `lane.RunBounded` | 301:20 |
| `runtime.laneStepFor` | 301:48 |
| `bounded.Err.Error` | 311:25 |
| `lane.Health` | 313:23 |

## Safety conclusion

- 레인 사이클 — 무편집 예상.
