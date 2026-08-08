# Branch Test Map: `ExitObserver.checkOutage`

Source: `internal/app/engine/exitloop.go` (767-804). AST 기준 분기 5 / 이탈 4 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:769` `lastObserved`가 zero → `startedAt` 사용 | 간접 — 첫 사이클부터 두절을 만드는 테스트 | no | yes |
| B2 | `:772` 경과 < 60s → 무동작 | `exitloop_test.go`의 두절 테스트가 임계 미만을 함께 돈다 | no | yes |
| B3 | `:775` 이미 알렸으면 다시 안 알린다 | 두절 래치 테스트 | no | yes |
| B4 | `:793` `Escalate` nil 또는 `AccountRef` 공백 → 승격 없음 | 간접 — 두 필드를 안 채우는 단위 테스트 다수 | no | yes |
| B5 | `:798` 승격이 오류 → `logErr` | **없음** — `EscalateOperatingMode`가 오류를 돌려주는 테스트가 없다 | no | no |

## 없는 것: 두 번의 `Notify`가 같은 사이클에서 연달아 일어난다는 단언

B4를 통과하는 경로에서 `Notify`가 **두 번** 불린다는 것을 단언하는 테스트가 없다.
프로덕션은 B4를 통과하지만 단위 테스트 대부분은 `Escalate`를 안 채워 **B4에서 나간다** —
그래서 두 번째 도달 경로가 테스트에서 아예 실행되지 않는다.

`a074_notification_wiring_test.go`가 배선을 단언하지만 두절 사이클 전체를 돌지는 않는다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | `Escalate`·`AccountRef`·`Announcer` 전부 채우고 응답하지 않는 publisher로 두절 사이클 1회 | `Notify` 2회, 사이클 체류 ≤ 2 × `Interval()`. 현행은 68s |
| R2 | 같은 조건에서 `EscalateOperatingMode`가 오류 (B5) | `logErr` 한 줄, 사이클 계속 |

R1이 a092가 H1에 대해 지는 계약이다.
