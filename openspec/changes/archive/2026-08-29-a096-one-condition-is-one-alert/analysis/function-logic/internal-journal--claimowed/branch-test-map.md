# Branch Test Map: `claimOwed`

측정: `go test -covermode=set ./internal/journal/...` — GREEN 75.0%.
이 함수는 a096 2판이 만들었으므로 RED(base `ec29dc72`)에 존재하지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch state` 진입 `:251` | 모든 claim 테스트 | 없음 | — |
| B2 | `case AlertPending` `:252` | `TestClaimingAFreshAlertOwesDelivery`, `TestClaimingAnUndeliveredAlertStillOwesDelivery` | 없음 | 진입 |
| B3 | `case Delivered/Acknowledged` `:255` | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | 없음 | 진입 |
| B4 | `remindAfter <= 0` — 재무장 금지 `:256` | `TestAZeroReminderWindowNeverReArms` | 없음 | 진입 |
| B5 | 종결됐는데 파싱 가능한 시각이 없다 `:260` | 없음 — 정상 mutator로는 만들 수 없다(아래) | 없음 | 미진입 |
| B6 | 종결 시각이 현재보다 미래다 → fail-open 재무장 `:266` | `TestASettledStampInTheFutureStillOwesDelivery` | 없음 | 진입 |
| B7 | reminder 창 안이다 → 억제 `:280` | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing`, `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | 없음 | 진입 |
| B8 | `default` — 모르는 상태를 owed+재무장 `:284` | `TestAnUnrecognisedAlertStateOwesDelivery` (PENDING 복귀와 완료 표시 포함) | 없음 | 진입 |

창을 넘긴 재무장 경로(B7 아래의 `return true, true`)는
`TestClaimingADeliveredAlertPastTheWindowIsReArmed`와
`TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` 후반부가 덮는다.

## B5에 대한 판단

`delivered_at`도 `acknowledged_at`도 파싱되지 않는 종결 행은 이 코드베이스의 mutator로는
만들 수 없다 — `MarkAlertDelivered`와 `AcknowledgeAlert`가 둘 다 자기 시각을 함께 쓴다.
외부 편집이나 미래의 마이그레이션만이 그런 행을 만들 수 있고, 그때의 안전한 답이 "보낸다"다.
테스트로 만들려면 SQL을 직접 써야 하며, 그렇게 만든 행은 이 함수가 아니라 그 SQL을
검증한다. `not-applicable`: 정상 경로로 도달 불가, 기본값이 안전 방향.
