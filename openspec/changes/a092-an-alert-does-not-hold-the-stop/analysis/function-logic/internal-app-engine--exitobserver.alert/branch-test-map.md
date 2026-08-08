# Branch Test Map: `ExitObserver.alert`

Source: `internal/app/engine/exitloop.go` (1600-1607). AST 기준 분기 2 / 이탈 1 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1601` `Alerts == nil` → 조용히 반환 | 간접 — `Alerts`를 안 채우는 exit 루프 단위 테스트 다수 | no | yes |
| B2 | `:1604` `Notify`가 오류 → `logErr` 한 줄 | **없음** — `Notify`가 오류를 돌려주는 exit 루프 테스트가 없다 | no | no |

## 분기가 아니라 시간이 비어 있다

분기 자체는 사소하다. **없는 것은 시간에 대한 단언이다.** `Notify`가 돌아오기까지
이 함수가 붙잡혀 있다는 사실을 단언하는 테스트가 exit 루프 쪽에 하나도 없다.
그래서 `Notify`의 예산이 10초에서 34초로 바뀌어도 이 패키지의 테스트는 깨지지 않는다.

대조군: `internal/obs`에는 예산을 아는 테스트가 있다 —
`TestPersistentDeliveryFailureBlocksEntries`(`obs_test.go:384`)는 재시도 소진을,
`TestNotifierIsConcurrencySafe`(`:620`)는 `deliver`의 `n.mu` 직렬화를 단언한다.
**두 성질 다 호출자 쪽에서는 관측되지 않는다.**

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 응답하지 않는 publisher를 주입하고 `ObserveOnce` 1회 | 사이클이 관측 주기(5s) 안에 반환한다 — 현행은 최대 34초 붙잡힌다 |
| R2 | 같은 조건에서 손절 판정이 걸린 포지션을 알림 뒤에 하나 더 둔다 | 두 번째 포지션의 제출이 첫 알림에 밀리지 않는다 |
| R3 | `Notify`가 오류를 돌려줌 (B2) | `logErr` 한 줄, 사이클은 계속 |
| R4 | `Alerts == nil` (B1) | 무동작, 체류 0 |

R1·R2가 이 change의 계약이다. 두 시나리오 모두 **현재 테스트가 존재하지 않는다.**
