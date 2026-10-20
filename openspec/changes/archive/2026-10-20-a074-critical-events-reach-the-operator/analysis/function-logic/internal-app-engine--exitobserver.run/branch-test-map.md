# Branch Test Map: `ExitObserver.Run`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 사이클이 주기마다 실행되고 실패한 사이클이 error 로그로 남는다 | `TestRunReportsAFailedCycle` | **yes** (M1) | pass |
| B2 | ctx가 이미 취소됨 → 즉시 반환 | `TestRunReturnsTheContextError` | no | pass |
| B3 | sleep 중 취소 → 반환 | `TestRunReturnsTheContextError` | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 침묵 | 성공한 사이클은 아무 줄도 남기지 않는다 | `TestRunSaysNothingAboutASuccessfulCycle` | no | pass |
| 등급 | 사이클 실패만으로는 알림도 운영 모드 강화도 없다 | `TestAFailedCycleAloneRaisesNoAlert` | no | pass |
| 반환 | 사이클 실패로 `Run`이 반환하지 않는다 | `TestRunKeepsGoingAfterAFailedCycle` | no | pass |

**M1의 RED는 실제 관측이다** (2026-08-04): `reportCycle`의 `o.logErr` 호출을 지우자
`TestRunReportsAFailedCycle`이 "cycle failure lines = 0, want one"으로 실패했다.
