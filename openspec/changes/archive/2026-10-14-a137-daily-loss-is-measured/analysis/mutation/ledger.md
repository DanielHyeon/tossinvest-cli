# a137 변이 원장

- 사본: 연결 워크트리 `df5dc288` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 44 시험 GREEN)

| 변이 | 파일 | 시험 | 결과 |
|---|---|---|---|
| M1 max → 호출자만 | `internal/execgw/a137_daily_loss.go` | `TestA137TheLedgerLossRefusesAZeroCallerAndEscalates` | CAUGHT |
| M2 측정 실패 → 호출자 값 낙하 | `internal/execgw/a137_daily_loss.go` | `TestA137AMeasurementFailureNeverFallsBackToTheCaller` | CAUGHT |
| M3 무효 호출자 값 대체 | `internal/execgw/a137_daily_loss.go` | `TestA137AnInvalidCallerValueIsNotReplaced` | CAUGHT |
| M4 이익 상쇄 | `internal/dailyloss/measure.go` | `TestA137ProfitsDoNotOffsetLosses` | CAUGHT |
| M5 시장 현지일 귀속 | `internal/dailyloss/measure.go` | `TestA137AttributionIsTheOneAccountDay` | CAUGHT |
| M6 조정 종결 무조건 제외 | `internal/dailyloss/measure.go` | `TestA137AnAdjustmentCloseWithADispatchedEngineSellIsUnmeasurable` | CAUGHT |
| M7 성과 부재 무시 | `internal/dailyloss/measure.go` | `TestA137AManagedCloseWithoutAnOutcomeIsUnmeasurable` | CAUGHT |
| M8 토큰 신선도 무시 | `internal/execgw/a137_daily_loss.go` | `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` | CAUGHT |
| M9 일 키 확인 삭제 | `internal/execgw/a137_daily_loss.go` | `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` | CAUGHT(build) |
| M10 다른 Guardian 토큰 허용 | `internal/execgw/a137_daily_loss.go` | `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` | CAUGHT |
| M11 위험 축소가 측정 | `internal/execgw/riskguardian.go` | `TestA137RiskReductionNeverMeasures` | CAUGHT |
| M12 IssueEntry 바닥 미적용 | `internal/execgw/riskguardian.go` | `TestA137TheLedgerLossRefusesAZeroCallerAndEscalates` | CAUGHT |
| M13 측정 실패 상세 미교체 | `internal/execgw/a137_daily_loss.go` | `TestA137AMeasurementFailureNeverFallsBackToTheCaller` | CAUGHT |
| M14 강화 생략 | `internal/app/engine/strategy_first_leg_admission.go` | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault` | CAUGHT |
| M15 Stage 검사 삭제 | `internal/app/engine/strategy_first_leg_admission.go` | `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` | CAUGHT |
| M16 precheck 원인 버림 | `internal/app/engine/strategy_first_leg_admission.go` | `TestA137APrecheckRefusalKeepsItsCause` | CAUGHT |
| M17 차단을 오류로 | `internal/app/engine/strategy_market_handoff_delivery.go` | `TestA137ADailyLossModeBlockEndsTheCycleQuietly` | CAUGHT |
| M18 건너뛴 범위 보고 삭제 | `internal/app/engine/strategy_market_handoff_delivery.go` | `TestA137AMixedCycleStillReportsItsSkippedScopes` | CAUGHT |
| M19 다른 trigger 도 조용히 | `internal/execgw/modegate.go` | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock` | CAUGHT |
| M20 다른 latch 와 함께여도 조용히 | `internal/execgw/a137_daily_loss.go` | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock` | CAUGHT |
| M21 펜스 원인 버림 | `internal/execgw/a137_daily_loss.go` | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` | CAUGHT |
| M22 checkEntry 표지 삭제 | `internal/execgw/gateway.go` | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` | CAUGHT |
| M23 관측 표지 삭제 | `internal/execgw/strategy_entry_gate_authority.go` | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock` | CAUGHT |
| M24 원장 보호 주문 무시 | `internal/journal/a137_daily_loss_rows.go` | `TestA137AnOverlappingProtectionOrderIsSeen` | CAUGHT |
| M25 원장 형식 결함 무시 | `internal/journal/a137_daily_loss_rows.go` | `TestA137AMalformedClosedAtMakesTheReadFail` | CAUGHT |
| M26 원장 NOT_DISPATCHED 포함 | `internal/journal/a137_daily_loss_rows.go` | `TestA137AnOverlappingDispatchedEngineSellIsSeen` | CAUGHT |
| M27 원장 조정 종결 decimal 0 → 문자 비교 | `internal/journal/a137_daily_loss_rows.go` | `TestA137AnAdjustmentToDecimalZeroIsAnAdjustmentClose` | CAUGHT |
| M28 환산 권한 없이 1:1 | `internal/execgw/a137_daily_loss.go` | `TestA137AForeignLossConvertsOnlyThroughThePairedAuthority` | CAUGHT |
| M2c 살아 있는 문 바닥 미적용 | `internal/execgw/riskguardian_qfinal.go` | `TestA137ThePrecheckDoorRefusesAMeasuredLossAtTheLimit` | CAUGHT |
| M1b 창 롤오버 삭제 | `internal/execgw/a137_daily_loss.go` | `TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead` | CAUGHT |
| M1c 읽기 여유 음수 | `internal/execgw/a137_daily_loss.go` | `TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead` | CAUGHT |
| M3b 펜스 단일 latch 무시 | `internal/execgw/a137_daily_loss.go` | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` | CAUGHT |
| M29 살아 있는 문 FX 없음 | `internal/execgw/riskguardian_qfinal.go` | `TestA137ThePrecheckDoorConvertsAUSLossOnlyWithThePairedAuthority` | CAUGHT |
| M30 측정 불가 표지 미부착 | `internal/execgw/a137_daily_loss.go` | `TestA137ThePrecheckDoorRefusesAnUnmeasurableDayAsTyped` | CAUGHT |
| M31 루프가 측정 불가를 결함으로 | `internal/app/engine/strategy_market_handoff_delivery.go` | `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` | CAUGHT |
| M32 승인의 측정 불가 갈래 삭제 | `internal/app/engine/strategy_first_leg_admission.go` | `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` | CAUGHT |
| M33 승인이 토큰을 버림 | `internal/app/engine/strategy_first_leg_admission.go` | `TestA137AdmissionCarriesTheMeasuredTokenIntoThePrecheck` | CAUGHT |
| M38 측정 불가 계수 미증가 | `internal/app/engine/a137_daily_loss_notice.go` | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay` | CAUGHT |
| M39 알림이 주기마다 | `internal/app/engine/a137_daily_loss_notice.go` | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay` | CAUGHT |
| M40 전달 자리 기록 삭제 | `internal/app/engine/strategy_market_handoff_delivery.go` | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay` | CAUGHT |
| M41 원인 판정 재실행 삭제 | `internal/execgw/a137_daily_loss.go` | `TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked` | CAUGHT |
| M42 주기 ctx 기록자 미부착 | `internal/app/engine/strategy_entry_supervisor.go` | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay` | CAUGHT |
| M43 생산 알림 배선 삭제 | `internal/app/engine/strategy_entry_supervisor.go` | `TestA137TheProductionSupervisorCarriesTheNotice` | CAUGHT |
| M44 표지를 다른 파일에서 지어냄 | `internal/app/engine/strategy_first_leg_admission.go` | `TestA137TheQuietMarkersAreMadeOnlyInTheirOwnFile` | CAUGHT |

이력: 1회차 `a1cfe0c2` 26 CAUGHT · M16 NOT-REACHED · M20 SURVIVED(→ 4fc5d4a4). 2회차 `4fc5d4a4` 28/28. 3회차 `c83241ff` 35 CAUGHT · M15 · M17 NOT-REACHED.
4회차 `3d48c22b` 37/37. 5회차(재확인 수리, 격리 사본 `b88d9f17`) 43 CAUGHT · M13 NOT-REACHED(P2-N1 로 대상 줄 바뀜). 6회차 `df5dc288`: **44/44 CAUGHT**.
리뷰어 판정 동등 M6 · M10 · M27b 는 동등으로 둠(review.md §6). `dailyLossFloor` 의 CompareDecimal 오류 갈래(B3)는 도달 불가 방어 갈래.
