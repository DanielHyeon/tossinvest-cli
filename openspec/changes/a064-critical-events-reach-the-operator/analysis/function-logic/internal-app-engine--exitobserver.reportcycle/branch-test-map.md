# Branch Test Map: `ExitObserver.reportCycle`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 성공한 사이클은 조기 반환하고 아무 줄도 남기지 않는다 | `TestRunSaysNothingAboutASuccessfulCycle` | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 보고 | 실패한 사이클이 error 등급 줄을 남긴다 | `TestRunReportsAFailedCycle` | **yes** (M1) | pass |
| 등급 | 사이클 실패가 알림·운영 모드를 건드리지 않는다 | `TestAFailedCycleAloneRaisesNoAlert` | no | pass |

**M1의 RED는 실제 관측이다** (2026-08-04): `o.logErr` 호출을 지우자
`TestRunReportsAFailedCycle`이 "cycle failure lines = 0, want one"으로 실패했다.
