# Branch Test Map: `ExitObserver.Run`

Source: `internal/app/engine/exitloop.go` (353-363). AST 기준 분기 3 / 이탈 2 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:354` 루프가 반복한다 | 간접 — `ObserveOnce`를 직접 부르는 테스트가 대부분이라 `Run` 자체를 도는 테스트는 드물다 | no | yes |
| B2 | `:355` `ctx` 취소 → 즉시 반환 | `exitloop_test.go`의 취소 경로 | no | yes |
| B3 | `:359` `Sleep`이 취소 → 반환 | 같은 위 | no | yes |

## 없는 것은 분기가 아니라 시간이다

세 분기 다 사소하고, **주기에 대한 단언이 하나도 없다.** `ObserveOnce`가 34초 걸려도
B1~B3은 전부 같은 값을 돌려주므로 이 함수의 테스트는 침묵한다.

`o.clk`가 주입 가능하므로(`clock.Clock`) **가짜 시계로 체류를 측정하는 테스트는 쓸 수
있다.** 오늘 그런 테스트가 없을 뿐이다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 응답하지 않는 publisher + critical 알림 1건이 걸리는 포지션 → `ObserveOnce` 1회 | 체류 ≤ `Interval()`(5s). 현행은 34s |
| R2 | 같은 조건에서 두절 사이클(P4 + P1a 둘 다 발동) | 체류 ≤ 2 × 5s. 현행은 68s |

R1·R2가 a092의 계약이고, 둘 다 **현재 존재하지 않는다.**
