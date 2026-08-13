# Branch Test Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (183-324)
- 이 change 가 편집한 분기는 **B17·B18** 이다. 나머지 행은 기존 고정을 그대로 옮긴 것이며
  RED/GREEN 열은 **이 change 에서 관측했는가**를 뜻한다 — 기존 분기는 `no/no` 가 정상이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil ctx 로 들어온다 | 간접 — 모든 CLI 경로가 cobra ctx 를 준다 | no | no |
| B2 | journal 경로 해석 실패 | — (미고정: 환경 이상, 이 change 밖) | no | no |
| B3 | lock 을 이미 누가 잡고 있다 | `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled` `engine_test.go:104` · 해제는 `TestTheLockIsReleasedWhenTheCommandReturns` | no | no |
| B4 | 조립이 인터록 아닌 이유로 실패 | `TestANonInterlockFailureIsReportedAsItself` `engine_test.go:194` | no | no |
| B5 | 인터록 절이 미충족 | `TestAnUnmetInterlockIsEnumerated` `engine_test.go:170` | no | no |
| B6 | 미충족 절을 열거한다 | `TestAnUnmetInterlockIsEnumerated` `engine_test.go:170` | no | no |
| B7 | 게이트 OFF | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` `engine_test.go:147` | no | no |
| B8 | verify lock 경로를 해석했다 | `TestAFreshVerifyRunLockRefusesTheStart` `engine_test.go:208` | no | no |
| B9 | verify lock 이 신선/오래됨 | `TestAFreshVerifyRunLockRefusesTheStart` · `TestAStaleVerifyRunLockDoesNotRefuse` `engine_test.go:231` | no | no |
| B10 | proc instance 를 읽어 마커에 싣는다 | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` `a102_ready_wiring_test.go:34` | no | no |
| B11 | 마커를 못 잡았다 (거절 아님) | — (미고정: 기존 강등, 이 change 밖) | no | no |
| B12 | 마커를 잡았다 | `TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter` `engine_test.go:253` | no | no |
| B13 | 루프 조립 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` `engine_runtime_branch_test.go:49` | no | no |
| B14 | policy command service 실패 | — (미고정) | no | no |
| B15 | policy command server 실패 | — (미고정) | no | no |
| B16 | policy runtime server 실패 | — (미고정) | no | no |
| B17 | projection 이 정상적으로 섰다 → defer Close 등록 (**이 change 가 만든 분기**) | `TestASucceedingProjectionIsStillServedAndClosed` `a108_the_engine_outlives_its_read_endpoint_test.go:264` | yes (뮤테이션 M4·M5) | yes |
| B18 | projection 기동 실패 → 강등 + durable critical 알림 + 루프 계속 (**이 change 가 만든 분기**) | `TestAFailedStrategyProjectionDoesNotStopTheEngine` `a108_the_engine_outlives_its_read_endpoint_test.go:155` · `TestTheDegradedBootLeavesADurableCriticalAlert` `:181` · `TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne` `:226` · `TestTheDegradedBootStillHoldsTheJournalFlock` `:301` · `TestTheDegradedBootPublishesReadyOnlyAfterRecovery` `:335` · `TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched` `:376` | yes | yes |
| B19 | alert 운영 표면 조립 실패 | — (미고정) | no | no |
| B20 | alert control server 실패 | — (미고정) | no | no |

### B18 이 지는 여섯 가지 (한 행에 접힌 이유는 분기가 하나이기 때문이다)

| 성질 | 테스트 | 죽이는 뮤테이션 |
|---|---|---|
| 엔진이 안 죽는다 | `TestAFailedStrategyProjectionDoesNotStopTheEngine` | M1 |
| 강등이 durable critical 알림을 남긴다 | `TestTheDegradedBootLeavesADurableCriticalAlert` | M1·M2 |
| 실행마다 알림, 실행 안에서는 한 번 | `TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne` | M1·M2·M3 |
| journal flock 싱글턴 불변 | `TestTheDegradedBootStillHoldsTheJournalFlock` | M1·M7 |
| a102 ready 발행 시점 불변 | `TestTheDegradedBootPublishesReadyOnlyAfterRecovery` | M1·M6 |
| automation interlock 평가 순서 불변 | `TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched` | M5 |

## RED 을 실제로 관측한 순서 (B18)

1. 무행위 리팩터로 `engineStrategyProjectionStart` seam 을 넣었다. 기존 574건 전부 GREEN
   유지 — 동작이 안 바뀌었다는 증거다.
2. 그 seam 에 사고 당일 오류(`stale endpoint is incomplete`)를 주입하는 테스트 4건을 썼다.
   **4건 전부 실패**했고, 실패 메시지는 전부
   `engine run = strategy projection runtime: stale endpoint is incomplete` 였다 —
   즉 현재 코드가 그 오류를 그대로 돌려주고 있었다(=exit 1).
   같은 실행에서 대조군 `TestASucceedingProjectionIsStillServedAndClosed` 와
   게이트 순서 테스트는 **통과**했다. harness 가 루프까지 실제로 간다는 증거다.
3. 강등을 구현하니 4건이 GREEN 이 됐고 전체 스위트는 574 → 582 로 늘었다.

## 미고정으로 남긴 분기 (사유)

B2·B11·B14·B15·B16·B19·B20 은 **이 change 가 편집하지 않은 기존 실패 경로**다.
전부 「조립 실패 → return」의 같은 모양이고, 고정하려면 각각의 생성자를 실패시키는
harness 가 필요하다. a108 의 범위는 B17·B18 이므로 선언된 생략이다.
