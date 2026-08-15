# Branch Test Map: `runEngineRun`

편집 **전** 상태다(a109 base `016da624`). B15·B16·B20 이 이 change 가 여는 세 분기이고,
그 세 줄만 RED 로 고정한 뒤 GREEN 으로 바꾼다. 나머지는 기존 핀의 현황 기록이다.

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
| B15 | **policy command server 기동 실패 → 오늘 엔진 사망** | a109 §2.1 RED → §2.2 GREEN (`cmd/tossctl/a109_*_test.go`) | pending | pending |
| B16 | **policy runtime server 기동 실패 → 오늘 엔진 사망** | a109 §2.1 RED → §2.2 GREEN (`cmd/tossctl/a109_*_test.go`) | pending | pending |
| B17 | projection 성공 경로만 Close 를 등록한다 | `TestASucceedingProjectionIsStillServedAndClosed` | no | yes |
| B18 | projection 기동 실패는 강등이고 원장에 안 쓴다 | `TestAFailedStrategyProjectionDoesNotStopTheEngine` · `TestTheDegradedBootWritesNoUndeliveredOutboxRow` · `TestTheDegradedBootDoesNotWaitForTheNotifier` | yes (a108) | yes |
| B19 | alert 운영 표면 조립 실패는 fatal 유지 | 미고정 (endpoint 아님 — 이 change 밖) | no | no |
| B20 | **alert control server 기동 실패 → 오늘 엔진 사망** | a109 §2.1 RED → §2.2 GREEN (`cmd/tossctl/a109_*_test.go`) | pending | pending |
