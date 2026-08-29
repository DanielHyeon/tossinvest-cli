# Branch Test Map: `MeasureFirstSighting`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 두 심볼이 섞인 슬라이스 | `TestASightingWithNoRankIsUnmeasured`(L753) | — (기존 동작) | yes |
| B2 | 요약이 없는 후보 | `TestSeenLateWithoutACandidateCannotKnowWhenWeFirstSaw` | — (기존 동작) | yes |
| B3 | 저장된 위치가 없다 | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps`(네 가지 nothing이 서로 다른 사유) | — (기존 동작) | yes |
| B4 | 거부 전에 읽은 것을 싣는다 | `TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate`(4 of 100 보고) · `TestATruncatedReadingsPositionIsNotAPercentile` | yes | yes |
| B5 | instant 없는 저장 위치 | `TestTheFirstSightingBoundaryIsTheStalenessTTL` | — (기존 동작) | yes |
| B6 | 다른 생명의 위치 | `TestTheFirstSightingBoundaryIsTheStalenessTTL` · `TestAPreviousLifesReadingsAreNotThisCandidatesFirstSighting` | — (기존 동작) | yes |
| B7 | 소스가 직전 읽기를 갖고 있지 않았다 | `TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate` · `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps` · `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan` · `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows` | yes | yes |
| B8 | 요청 행 수가 기록되지 않은 위치 | `TestAPositionWithNoRecordedRequestIsRefusedUnderItsOwnReason` · `TestTheOneRowReadingWithNoRecordedRequestIsTheCaseThatMattered`(백분위 0이 모든 임계를 통과한다는 산술 포함) | yes (I3의 원래 결정은 이것을 `Measured`로 뒀고, F4가 뒤집었다) | yes |
| B9 | 100 요청 3 도착의 1위, 그리고 1행 읽기 | `TestATruncatedReadingsPositionIsNotAPercentile`(66.7이 나오면 실패) · `TestTheOneRowReadingIsCaughtByTheSameRefusal` · `TestATruncatedReadingReachesTheVerdictAsTruncated`(배선) | yes | yes |
| (꼬리) | 자격을 전부 갖춘 위치는 측정된다 | `TestAShortListThatIsNotTruncatedIsStillMeasurable`(30/30) · `TestASessionStartDoesNotStampThePanelAsSeenLate`(두 번째 tick) · `TestARePromotionAfterExpiryIsQualifiedByTheReadingThatSawTheSymbolReturn` | yes | yes |
| (비율 없음) | 이 함수가 패널의 비율을 세지 않는다 | `TestTheRefusalCountsNoRatioOfThePanel`(AST: 0 이외의 숫자 리터럴 금지, 나눗셈 금지) | yes | yes |
