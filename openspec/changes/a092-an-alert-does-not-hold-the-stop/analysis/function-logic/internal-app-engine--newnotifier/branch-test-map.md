# Branch Test Map: `newNotifier`

Source: `internal/app/engine/exitwiring.go` (71-81). AST 기준 **분기 0** / 이탈 1 /
calls 0 / defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없음 — 구조체 리터럴 1개를 돌려준다 | 간접 — `gateway.go:280`을 지나는 조립 테스트 전부 | no | yes |

## 분기가 없으므로 빈 것은 *값*에 대한 단언이다

이 함수가 **채우지 않은** 필드를 단언하는 테스트가 없다. `Attempts`·`RetryDelay`가
비어 있다는 사실은 어느 테스트도 보지 않고, 그래서 34초 예산은 테스트에 나타나지 않는다.

가장 가까운 대조군은 `a074_notification_wiring_test.go:167`
`TestAWiredPublisherDeliversTheCriticalAlert`인데, 그것은 **publisher가 배선되는지**만
단언하고 예산은 보지 않는다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | `newNotifier(...)`의 `Attempts` | 명시된 값(현행 RED: 0) |
| R2 | `newNotifier(...)`의 `RetryDelay` | 명시된 값(현행 RED: 0) |
| R3 | 세 값으로 계산한 critical 최악 예산 | ≤ `DefaultExitObservationInterval`(5s) (현행 RED: 34s) |

R3이 a092의 계약이다 — **예산을 상수 하나가 아니라 세 필드의 합으로 단언한다.**
