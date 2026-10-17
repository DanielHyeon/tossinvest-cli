# Function Logic Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go`
- AST evidence: `ast.json` — **편집 전**, :270–340, 분기 11, source_sha256 `5791a31af9d2…`, 추출 HEAD `b01e0cd0`(작업 트리 — 파일 무편집).
- Risk scan: `risk-pattern-report.md`
- 편집 목적(25라운드 보이스 A #2): 만료 임차 인수(`claim.Stole`) 줄의 이벤트 타입을 `EventAlertClaimHeld` → `EventAlertClaimStolen`. 알림기(단위 ③)가 `claim_held` 를 정상 경합(INFO)으로 내렸으므로, 죽은 발송자 신호는 두 발송 경로 모두 `claim_stolen` 이 져야 함 — 같은 이름이 한쪽은 INFO(보유) · 한쪽은 WARN(탈취)이면 경보 규칙을 걸 수 없음. 분기 · 판정 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `alert` | 나열된 PENDING 행 | 배달 실행자 사이클 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:272) | — | — | (미실행) |
| B2 | `switch claim.Disposition` (:278) | — | — | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch`, `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B3 | `case journal.ClaimAcquired:` (:279) | — | — | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch`, `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B4 | `case journal.ClaimHeldElsewhere:` (:283) | — | — | (미실행) |
| B5 | `if d.reportHeld(alert.ID, claim.ExpiresAt)` (:292) | — | — | (미실행) |
| B6 | `default:` (:298) | — | — | (미실행) |
| B7 | `if claim.Stole` (:305) | — | — | (미실행) |
| B8 | `if d.Publisher == nil` (:314) | — | — | (미실행) |
| B9 | `} else` (:324) | — | — | (미실행) |
| B10 | `if perr != nil` (:331) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B11 | `if perr != nil` (:335) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.logf(obs.EventAlertClaimHeld, …, "an expired alert lease was taken over", …)` | 인수 기록 | 로그 | AST — 편집 대상(이벤트 타입만) |

## State mutations and fallbacks

- 편집 전과 같음 — 로그 이벤트 타입 한 곳.

## Safety conclusion

- Safe edit boundary: 인수 로그 줄의 이벤트 인자 하나.
- High-risk impact: 낮음(관측 이름) — 배달 판정 · 래치 · 승격 무변화.
