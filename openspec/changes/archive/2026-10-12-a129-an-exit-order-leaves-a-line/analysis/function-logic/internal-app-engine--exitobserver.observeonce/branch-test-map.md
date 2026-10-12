# Branch Test Map: `ExitObserver.ObserveOnce`

**편집 뒤 좌표**(분기 구조 불변, +14). 편집 전 표는 git 이력. 진입 실측: 편집 뒤 엔진 패키지 전체 커버리지(set, 2026-10-12) — 8 분기 블록 전부 count 1.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:462` 체결 감지 양보 — defer 가 이 return 에서도 돎 | `TestA129CTheFillFollowRunsOnAYieldedCycle` · `TestTheCycleYieldsToFillDetection` | yes (red-1) | yes |
| B2 | `:472` 작업 집합 오류 | 엔진 패키지 커버리지 count 1 — 시험 이름 미귀속, 무편집 | no | yes |
| B3 | `:476` 보유 0 — 전량 청산 뒤 체결 줄의 주 경로 | `TestA129AConfirmedStopAndItsFillsEachLeaveALine`(전량 체결 뒤 주기) · `TestAnAccountHoldingNothingIsNotInAnOutage` | yes | yes |
| B4 | `:489` 가격 읽기 실패 | `TestAFailedObservationHoldsTheJudgement` · `TestASustainedOutageBlocksEntriesAndAlertsOnce` | no | yes |
| B5 | `:498` 포지션 순회 — 판정이 따라가기보다 먼저 | `TestA129BTheFillFollowRunsAfterTheJudgementOfTheSameCycle` · `TestABaselineBreachProposesTheWholePosition` | yes | yes |
| B6 | `:500` 응답 없음 | `TestA090R1AZeroQuoteSiblingIsCountedAndLoggedAtNormalGrade` | no | yes |
| B7 | `:508` 임대 만료 | `TestA090R7AQuoteThatExpiredMidCycleIsCountedAsExpired` | no | yes |
| B8 | `:516` judge 오류 | `TestA090R16AJudgementThatEndsAtOnceStillCountsAsObserved` · `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease/B4_attach_failed`(judge 오류 주기에도 줄) | no | yes |

defer 의 위치는 변이 M8(defer → 직접 호출, 판정 앞에서 돎) CAUGHT 로 고정(`analysis/mutation/round4.log`). `Place` 패닉이 defer 를 지나 전파됨은
`TestA129KAPanicElsewhereInTheCycleIsNotSwallowed`.
