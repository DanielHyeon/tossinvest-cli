# Function Logic Map (편집 전): `runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- Source range: `252:1`–`274:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: 관측에 활성화가 말한 desired/effective 와 입력의 snapshot · evidence digest 를 싣는다. 레인 상태 호출(Offer · RunBounded) 순서 · 횟수 불변.

## Branches and early returns

- Exact AST return nodes: `258:3`, `273:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 256:2 | `if observation.Trigger != strategyworker.TriggerEnqueued {` |
| B2 | if | 268:2 | `if bounded.Err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 255:46 |
| `lane.Offer` | 255:67 |
| `lane.Health` | 257:24 |
| `lane.RunBounded` | 260:20 |
| `strategyFamilyLaneStep` | 260:48 |
| `bounded.Err.Error` | 269:25 |
| `lane.Health` | 271:23 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
