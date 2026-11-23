# a194 변이 원장 (tasks 4.1 · 4.3, 2026-11-23)

- 하네스: `run_mutations.py`(커밋됨). 사본: 스크래치 `a194-mut-<pid>`(git archive HEAD), 최종 판 시작 sha `e26831d5a1c875e1c9e115908b340e9f1b5a4386`, 무변이 대조군 GREEN(rc 0, FAIL 0), 한 번에 한 판, 원복은 원문 문자열을 다시 세어 1 회 확인.
- 닿음: 무변이 대조군의 커버(set 모드)에서 변이 원문이 있는 줄의 블록이 실행됐는가. 선택 시험 정규식은 하네스 `RUN`.
- 판 기록: 1판(5a547f80d, M1~M12 13개) → 구현 리뷰 r1 뒤 2판(f0465bdfd, +M3r·M7a·M7b·M12r·M19·M23, 19/19 — `run-r1.txt`) → gstack 수리 뒤 최종 3판(아래, +M24). 원문 `run.txt`, 구조화 `results.json`.

| id | 변이 | 자리 | 닿음 | 결과 | 죽인 시험 |
|---|---|---|---|---|---|
| M1 | 운영자 전용 판정을 항상 거짓 | reconcile_block_report.go:84 | yes | CAUGHT | `TestA194ADeliveredHoldingReportIsRemindedAfterTheWindow`, `TestA194AFailedHoldingReportIsRecordedOnTheNextObservation`, `TestA194ALateDeliveredReportIsStillTrue`, `TestA194ARoundWhoseRestLeftStartsAgainNextCycle`, `TestA194ASnapshotCauseIsOperatorOnlyByItsRelease`, `TestA194AnAccountWideBlockIsOneReport`, `TestA194AtMostThreeHoldingReportsPerCycle`, `TestA194EarlierExcludedHoldingsAreReportedFirst`, `TestA194NotificationsOffNeverRecordsACritical`, `TestA194OnlyAdoptionCandidatesAreReported`, `TestA194OperatorOnlyBlockedHoldingIsReportedOnTheFirstCycle`, `TestA194TheEarliestCoveringOperatorBlockIsNamed`, `TestA194TheReportComesAfterTheCyclesAdoptions`, `TestA194TheReportsCarryNoAccountQuantityOrPrice` |
| M2a | 유예 비교 경계(< → <=) | reconcile_block_report.go:87 | yes | CAUGHT | `TestA194ABlockTheTrackerNoLongerListsTakesTheGracePath`, `TestA194ASelfReleasingBlockIsReportedAfterTenMinutes`, `TestA194ASkippedCycleDoesNotResetTheGraceClock`, `TestA194TwoFailedCyclesAtSixMinutesDoNotBreakTheRun` |
| M2b | 유예 비교 방향 뒤집기(< → >) | reconcile_block_report.go:87 | yes | CAUGHT | `TestA194ABlockClearedInsideTheGraceAdoptsSilently`, `TestA194ABlockTheTrackerNoLongerListsTakesTheGracePath`, `TestA194ASelfReleasingBlockIsReportedAfterTenMinutes`, `TestA194ASkippedCycleDoesNotResetTheGraceClock`, `TestA194TheGraceClockIsClearedWhenTheHoldingIsNotExcluded`, `TestA194TwoFailedCyclesAtSixMinutesDoNotBreakTheRun`, `TestAdoptionIsSilentUnderReconcile` |
| M3 | 시계를 지우지 않기 | reconcile_block_report.go:101 | yes | CAUGHT | `TestA194TheGraceClockIsClearedWhenTheHoldingIsNotExcluded` |
| M4 | D4 조건 삭제 | reconcile_block_report.go:73 | yes | CAUGHT | `TestA194OnlyAdoptionCandidatesAreReported` |
| M5 | 차단 보고 key 에서 날짜 삭제 | reconcile_block_report.go:213 | yes | CAUGHT | `TestA194AnOperatorOnlyBlockIsReportedOncePerDay`, `TestA194TheDayIsTheKoreanCalendarDay` |
| M6 | 보유 보고 critical 앞에 메모리 래치 | reconcile_block_report.go:130 | yes | CAUGHT | `TestA194ADeliveredHoldingReportIsRemindedAfterTheWindow`, `TestA194AFailedHoldingReportIsRecordedOnTheNextObservation`, `TestA194ARoundWhoseRestLeftStartsAgainNextCycle`, `TestA194AnAccountWideBlockIsOneReport`, `TestA194EarlierExcludedHoldingsAreReportedFirst`, `TestA194OperatorOnlyBlockedHoldingIsReportedOnTheFirstCycle` |
| M7 | 차단 보고 래치를 Notify 앞에(결과 무시) | reconcile_block_report.go:214 | yes | CAUGHT | `TestA194AFailedBlockReportIsRetriedNextCycle`, `TestA194NoAlerterIsNotARecordedBlockReport` |
| M8 | KST 날짜를 UTC 날짜로 | reconcile_block_report.go:206 | yes | CAUGHT | `TestA194TheDayIsTheKoreanCalendarDay` |
| M9 | Blocks().Release 대신 EntryAllowed 사유(Permanent)로 판정 | reconcile_block_report.go:153 | yes | CAUGHT | `TestA194ASnapshotCauseIsOperatorOnlyByItsRelease` |
| M10 | 주기 상한 삭제 | reconcile_block_report.go:133 | yes | CAUGHT | `TestA194ARoundWhoseRestLeftStartsAgainNextCycle`, `TestA194AtMostThreeHoldingReportsPerCycle` |
| M11 | 새 보고를 adopt 앞으로 옮기기 | adoption.go:181 | yes | CAUGHT | `TestA194TheReportComesAfterTheCyclesAdoptions` |
| M12 | 순서를 포지션 id 만으로 | reconcile_block_report.go:107 | yes | CAUGHT | `TestA194EarlierExcludedHoldingsAreReportedFirst` |
| M3r | 덮는 블록 중 첫 것(가장 이른 Since 아님) | reconcile_block_report.go:156 | yes | CAUGHT | `TestA194TheEarliestCoveringOperatorBlockIsNamed` |
| M7a | 운영자 전용 본문에서 할 일 문장 삭제 | reconcile_block_report.go:171 | yes | CAUGHT | `TestA194OperatorOnlyBlockedHoldingIsReportedOnTheFirstCycle` |
| M7b | stale 본문에서 할 일 문장 삭제 | reconcile_block_report.go:175 | yes | CAUGHT | `TestA194ASelfReleasingBlockIsReportedAfterTenMinutes` |
| M12r | 알림기 없음을 보낸 것으로 셈 | reconcile_block_report.go:244 | yes | CAUGHT | `TestA194NoAlerterIsNotARecordedBlockReport` |
| M19 | 제목에서 label 대신 코드 | reconcile_block_report.go:190 | yes | CAUGHT | `TestA194TheTitlesUseTheInstrumentLabel` |
| M23 | 풀린 차단의 래치 key 유지 | reconcile_block_report.go:218 | yes | CAUGHT | `TestA194ABlockRaisedAgainTheSameDayIsReportedAgain` |
| M24 | 차례가 빈 주기에 새 차례를 시작하지 않음(gstack review 수리 되돌림) | reconcile_block_report.go:122 | yes | CAUGHT | `TestA194ARoundWhoseRestLeftStartsAgainNextCycle` |

합계: 20 변이, CAUGHT 20, SURVIVED 0, 닿지 않음 0.
M6 은 두 자리 변이(꺼짐 전용 래치를 켜짐에도 세우고 건너뜀 조건에서 꺼짐 전제를 뺌), M11 은 끝의 호출을 지우고 adopt 앞에 넣음.
구현 리뷰 r1 이 원장에만 남기라고 한 M16(`blockedRound` 미정리)은 이 원장에서 돌리지 않음 — 영향 미미(Manager 처분).
