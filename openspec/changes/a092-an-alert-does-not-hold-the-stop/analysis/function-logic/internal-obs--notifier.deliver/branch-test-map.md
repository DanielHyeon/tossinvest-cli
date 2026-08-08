# Branch Test Map: `Notifier.deliver`

Source: `internal/obs/notifier.go` (238-287). AST 기준 분기 10 / 이탈 2 / defers 1.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:245` `attempts <= 0` → 기본 3회 | `TestPersistentDeliveryFailureBlocksEntries` (`obs_test.go:384`) | no | yes |
| B2 | `:251` 시도 루프 | 같은 테스트 | no | yes |
| B3 | `:252` publisher 없음 → 즉시 `break` | `TestCriticalAlertIsDurableBeforeItIsSent` (`:353`) — 행이 PENDING·attempts 0으로 남는다 | no | yes |
| B4 | `:257` publish 성공 → `return true` `:261` | `TestAWiredPublisherDeliversTheCriticalAlert` (`a074_notification_wiring_test.go:167`) | no | yes |
| B5 | `:258` `MarkAlertDelivered` 오류 → 로그 | **없음** — 그런데 **프로덕션에서 6번 발생했다**(2026-08-05 `engine.log`, `journal: no such alert: N (or it is no longer pending)`) | no | no |
| B6 | `:264` `MarkAlertAttemptFailed` 오류 → 로그 | **없음** | no | no |
| B7 | `:267` `attempt < attempts` → 대기 | `TestPersistentDeliveryFailureBlocksEntries` (주입 시계) | no | yes |
| B8 | `:268` ctx 종료 → `break` | **없음** | no | no |
| B9 | `:278` Log 있음 → 소진 기록 | 간접 | no | yes |
| B10 | `:283` Gate 있음 → **`Gate.Block`** | `TestPersistentDeliveryFailureBlocksEntries` (`:384`) | no | yes |

## 무테스트 이탈 3개 중 하나는 실제로 발생했다

**B5**는 테스트가 없는데 실운영에서 6번 발생했다 — `EnqueueAlert`가 기존 DELIVERED 행의
id를 재사용하고(`outbox.go:131`) `MarkAlertDelivered`의 `state=PENDING` 술어(`:159`)가
그것을 거절하기 때문이다. **a089의 범위이고 이 change는 건드리지 않는다.**
여기 적는 이유는 그 6줄이 **이 change의 실측 표본이기도** 하기 때문이다 —
이벤트 줄과 그 로그 줄의 간격이 동기 체류 시간이다.

## 이 change가 요구하는 RED

| # | Scenario | 기대 |
|---|---|---|
| R14 | `deliver`가 호출자 goroutine이 아닌 곳에서 돈다 | B1~B10의 결과가 전부 동일 |
| R15 | 발송 중 두 번째 critical이 들어온다 | `n.mu` 직렬화 유지, 호출자는 대기하지 않는다 |
| R16 | B8 — ctx가 종료된 상태 | `break` 후 게이트 래치. 종료 중 알림이 무한 대기하지 않는다 |

**B5·B6·B8의 무테스트 상태는 이 change가 닫지 않는다** — B5·B6은 a089,
B8은 R16이 부분적으로만 덮는다. `not-applicable` 사유를 review에 남긴다.
