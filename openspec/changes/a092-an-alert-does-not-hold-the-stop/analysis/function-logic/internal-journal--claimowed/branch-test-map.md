# Branch Test Map: `claimOwed`

Source: `internal/journal/outbox.go` (269-315). AST 기준 분기 8 / 이탈 7 /
defers 0 / go_statements 0.

이 함수는 unexported이므로 테스트는 전부 `ClaimAlertForDelivery`를 통해 관측한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:275` switch 진입 | 아래 B2·B3·B8이 대표 | — | — |
| B2 | `:276` **PENDING → `(true, false)`** | `a096_claim_for_delivery_test.go TestClaimingAnUndeliveredAlertStillOwesDelivery:164` | no | yes |
| B3 | `:279` DELIVERED/ACKNOWLEDGED 진입 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing:50`, `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow:131` | no | yes |
| B4 | `:280` `remindAfter <= 0` → `(false,false)` | `TestAZeroReminderWindowNeverReArms:218`, `a097_rearm_is_a_new_episode_test.go TestARecordOnlyCallerDoesNotReArmADeliveredRow:243` | yes (a097) | yes |
| B5 | `:284` 날짜를 못 읽는 행 → fail-open | **없음** — `delivered_at`을 깨뜨려 넣는 테스트가 없다 | no | no |
| B6 | `:290` **시계 역행 → fail-open** | `a096b_round2_test.go TestASettledStampInTheFutureStillOwesDelivery:30` | yes (a096 라운드 2) | yes |
| B7 | `:304` 창 안 → `(false,false)` | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing:50` | no | yes |
| — `:307` | 창 지남 → `(true,true)` | `TestClaimingADeliveredAlertPastTheWindowIsReArmed:91` | no | yes |
| B8 | `:308` **미상 상태 → fail-open** | `TestAnUnrecognisedAlertStateOwesDelivery:188`, `a097_rearm_is_a_new_episode_test.go TestRecoveringAnUnknownStateIgnoresTheReminderWindow:209` | yes (a097) | yes |

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 새 RED가 없다. **B2 한 행이 17판 D0.3의 전제**이므로,
§6.0 R17-3이 그 전제를 동시성 조건에서 다시 관측한다.

미테스트 분기는 B5 하나다. `delivered_at`에 파싱 불가 문자열을 직접 UPDATE로
넣으면 주입 없이 테스트할 수 있으므로, **b5는 "불가능"이 아니라 "안 만들었다"**이다.
a092는 이 함수를 편집하지 않으므로 여기서 만들지 않는다
(`not-applicable`: 편집 대상 아님). 기록해 두고 넘어간다.
