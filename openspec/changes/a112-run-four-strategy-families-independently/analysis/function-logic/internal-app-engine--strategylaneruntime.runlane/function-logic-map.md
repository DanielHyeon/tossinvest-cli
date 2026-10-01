# Function Logic Map: `strategyLaneRuntime.runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- Source range: `260:1`–`287:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 새 필드는 이미 받은 값(활성화 · 입력)의 읽기다 — 레인 상태를 바꾸는 호출을 더하지 않았다.
- desired/effective 는 관측 시점 계산(판정 Q2 부가): `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith`, 변이 P06 CAUGHT. 거절 코드 기록은 같은 시험(켜진 KR 레인 REFUSED + ARBITRATION_SEAL_MISMATCH), 변이 P22 CAUGHT.

## Branches and early returns

- Exact AST return nodes: `270:3, 286:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 268:2 | 투입 거절(DISABLED · FULL) → 건강만 싣고 반환 |
| B2 | if | 281:2 | 유계 사이클 오류 → 실패 문장 |

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

## State mutations and fallbacks

- 레인 상태는 Offer · RunBounded 만 바꾼다(편집 전과 같음).

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
