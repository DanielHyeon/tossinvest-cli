# Branch Test Map: `strategyRuntimeReaderFor`

편집 **후** 상태다(a109 §2.3 GREEN).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ctx 없이 불린다 | 간접 — 프로덕션 호출자(`runHTTPAPI`)는 언제나 ctx 를 준다 | no | no |

**분기 밖 계약 핀**(이 함수가 실제로 지켜야 하는 것):

| 계약 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 냉부팅 순서: 엔진이 나중에 떠도 재시작 없이 회복 | `TestTheDaemonAttachesWhenTheEngineComesUpLater` | yes (NOT_CONFIGURED 고착) | yes |
| 가동 중 재시작(새 socket·새 토큰) 뒤 회복 | `TestTheDaemonReattachesAfterTheEngineRestarts` | yes (RUNTIME_UNAVAILABLE 고착) | yes |
| 부팅 1회 해석의 세 화면 값이 그대로 유지된다 | `TestADialFailureRendersUnavailableRatherThanNotConfigured` | no (a108 회귀 핀) | yes |
| 요청 경로가 dial 을 기다리지 않는다 | `TestTheRequestPathNeverWaitsForADial` | no (기전 핀, GREEN 에서 도입) | yes |
| 시도는 겹치지 않는다 (rate limit 을 끈 상태에서 측정) | `TestTheAttemptIsSingleFlight` | no (기전 핀) — 뮤테이션 M10 이 구판의 결함을 잡아 분리했다 | yes |
| 창 안에서는 다시 시도하지 않고 창 밖에서는 한다 | `TestTheAttemptIsRateLimited` | no (기전 핀) | yes |
| 운영 기본 간격은 30s | `TestTheProductionRedialIntervalIsThirtySeconds` | no | yes |
| 실패한 시도가 현재 화면 값을 흔들지 않는다 | `TestAFailedAttemptDoesNotClobberTheCurrentScreen` | no | yes |
| 보고는 상태 전이 시 1회 | `TestTheAttachmentReportsOnlyTransitions` | no | yes |
