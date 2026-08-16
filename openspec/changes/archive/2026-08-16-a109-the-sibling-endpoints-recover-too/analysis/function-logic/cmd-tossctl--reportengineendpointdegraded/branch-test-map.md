# Branch Test Map: `reportEngineEndpointDegraded`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Notifier 가 없으면 stderr 한 줄만 | 직접 핀 없음 — 조립 실패 배포의 최소 표면. non-nil 경로는 `TestAFailedSiblingEndpointDoesNotStopTheEngine` 이 지난다 | no | no |

**분기 밖 계약 핀**:

| 계약 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 강등이 사유를 stderr 로 말한다 | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | yes | yes |
| 보고가 **어느 표면인지** 말하고 남의 표면은 말하지 않는다 | `TestADegradedSiblingSaysWhichSurfaceIsMissing` | yes | yes |
| 잃는 것을 정직하게 적는다(격리 해제 포함) | `TestEveryDegradedEndpointNamesWhatItLost` | no — GREEN 에서 도입 | yes |
| 네 강등이 동시에 나도 넷 다 보고된다 | `TestEveryEndpointCanFailAtOnceAndTheLoopsStillRun` | yes | yes |
| 미전달 outbox 행 0 · 진입 게이트 미잠금 | `TestADegradedSiblingBootWritesNoUndeliveredOutboxRow` · `TestTheDegradedBootWritesNoUndeliveredOutboxRow` | yes | yes |
| 두 이벤트 타입 모두 critical rail 밖 | `TestTheDegradationEventsAreNotOnTheCriticalRail` | no — GREEN 에서 도입 | yes |
| 느린 publisher 가 부팅을 붙잡지 않는다 | `TestTheDegradedBootDoesNotWaitForTheNotifier` | no (a108 핀) | yes |
| 금지 3종의 이유 주석이 남아 있다 | `TestTheDegradationCommentStillCitesTheForbiddenThree` | no — GREEN 에서 도입 | yes |
