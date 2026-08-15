# Branch Test Map: `runEngineRun`

편집 **후** 상태다(a109 §2.1 RED → §2.2 GREEN).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ctx 없는 호출 | 간접 — 모든 CLI 경로가 지난다 | no | no |
| B2 | journal 디렉터리 해석 실패 | 미고정 (환경 이상) | no | no |
| B3 | 둘째 엔진이 flock 에 막힌다 | `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled` | no | yes |
| B4 | 인터록 아닌 조립 실패 | `TestANonInterlockFailureIsReportedAsItself` | no | yes |
| B5 | 미충족 절 열거 | `TestAnUnmetInterlockIsEnumerated` | no | yes |
| B6 | 절 한 줄씩 | `TestAnUnmetInterlockIsEnumerated` | no | yes |
| B7 | 게이트 OFF 거절 | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` | no | yes |
| B8 | verify lock 경로 해석 | `TestAFreshVerifyRunLockRefusesTheStart` | no | yes |
| B9 | 신선한 verify lock 거절 / 오래된 것은 통과 | `TestAFreshVerifyRunLockRefusesTheStart` · `TestAStaleVerifyRunLockDoesNotRefuse` | no | yes |
| B10 | proc instance 토큰을 마커에 싣는다 | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` | no | yes |
| B11 | 마커 실패는 거절이 아니다 | 미고정 (기존 강등) | no | no |
| B12 | 마커 보유·해제 | `TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter` | no | yes |
| B13 | 루프 조립 실패는 fail-closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | no | yes |
| B14 | policy command service 실패는 fatal 유지 | 미고정 (endpoint 아님 — 이 change 밖) | no | no |
| B15 | policy command server 가 서면 Close 를 등록한다 | `TestSucceedingSiblingEndpointsAreStillServedAndClosed` | no | yes |
| B16 | policy command 기동 실패는 **강등**이고 부팅은 계속된다 | `TestAFailedSiblingEndpointDoesNotStopTheEngine` · `TestADegradedSiblingSaysWhichSurfaceIsMissing` · `TestADegradedSiblingBootWritesNoUndeliveredOutboxRow` | **yes** (엔진 사망) | yes |
| B17 | policy runtime server 가 서면 Close 를 등록한다 | `TestSucceedingSiblingEndpointsAreStillServedAndClosed` | no | yes |
| B18 | policy runtime 기동 실패는 **강등** | 위 셋과 같음 | **yes** (엔진 사망) | yes |
| B19 | projection 성공 경로만 Close 를 등록한다 | `TestASucceedingProjectionIsStillServedAndClosed` | no | yes |
| B20 | projection 기동 실패는 강등이고 원장에 안 쓴다 | `TestAFailedStrategyProjectionDoesNotStopTheEngine` · `TestTheDegradedBootWritesNoUndeliveredOutboxRow` | yes (a108) | yes |
| B21 | alert 운영 표면 조립 실패는 fatal 유지 | 미고정 (endpoint 아님 — 이 change 밖) | no | no |
| B22 | alert control server 가 서면 Close 를 등록한다 | `TestSucceedingSiblingEndpointsAreStillServedAndClosed` | no | yes |
| B23 | alert control 기동 실패는 **강등** | `TestAFailedSiblingEndpointDoesNotStopTheEngine` · `TestEveryEndpointCanFailAtOnceAndTheLoopsStillRun` | **yes** (엔진 사망) | yes |

**분기 밖 계약 핀**:

| 계약 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 네 표면이 동시에 없어도 루프는 돈다 | `TestEveryEndpointCanFailAtOnceAndTheLoopsStillRun` | yes | yes |
| 강등 기동도 journal flock 을 쥔다 | `TestTheDegradedSiblingBootStillHoldsTheJournalFlock` | yes | yes |
| lock.Release defer 가 모든 endpoint Close 보다 먼저 등록된다 | `TestTheJournalLockIsReleasedAfterEveryEndpointClose` | no — §2.4 정적 핀 | yes |
