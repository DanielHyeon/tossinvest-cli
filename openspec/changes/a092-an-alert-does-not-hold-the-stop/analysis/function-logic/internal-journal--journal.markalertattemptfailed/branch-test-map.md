# Branch Test Map: `Journal.MarkAlertAttemptFailed`

Source: `internal/journal/outbox.go` (352-363). AST 기준 분기 1 / 이탈 2 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:359` `ExecContext` 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| — `:362` | 실패 누적, 행은 `PENDING` 유지 | `outbox_test.go TestFailedAttemptsAccumulateAndTheRowStaysPending:75` | no | yes |
| — `:362` | 없는 id → `ErrAlertNotFound` | `outbox_test.go TestMarkingANonPendingAlertIsRefused:161` | no | yes |
| — `:362` | 재무장된 행에 다시 기록된다 | `a096_claim_for_delivery_test.go TestClaimingADeliveredAlertPastTheWindowIsReArmed:122` | no | yes |
| — `:362` | 재무장이 `attempts`를 0으로 되돌린다 | `a097_rearm_is_a_new_episode_test.go TestReArmingResetsTheAttemptCount:89` | yes (a097) | yes |

AST 분기는 1개이고 아래 네 행은 같은 이탈점(`:362`)의 서로 다른 결과다.
**AST에 없는 분기 ID를 만들지 않는다.**

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 이 함수에 대한 새 RED는 없다. 그러나 **이 함수를 부르지 않는
경로**에 대해서는 진다:

- **R17-4** — `n.Publisher == nil`일 때도 이 함수가 불려서 `attempts`가 늘고
  `last_error`가 남는지. HEAD `Flush:442-444`는 부르지 않는다.
- **R17-6** — 배달 루프의 한 주기가 한 행에 대해 이 함수(또는
  `MarkAlertDelivered`)를 **정확히 한 번** 부르는지.

두 테스트 모두 이 함수의 분기가 아니라 호출자를 관측하므로 위 표에 행을 만들지
않는다.

B1은 SQLite 실패 주입이 필요하고 a092의 범위가 아니다
(`not-applicable`: 이 change는 이 함수를 편집하지 않는다).
