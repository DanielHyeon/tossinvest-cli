# Branch Test Map: `officialRanking.Read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 일반 오류는 그대로 래핑 | `internal/candidatesrc` 오류 경로(`TestARateLimitedRankingIsReportedAsOne`의 fake err) | — (기존 동작) | yes |
| B2 | 429는 `candidate.ErrRateLimited`, 그리고 기억은 손대지 않는다 | `TestARateLimitedRankingIsReportedAsOne` · `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` | yes (F1 probe) | yes |
| B3 | 도착한 행마다 요청 수와 3-상태를 싣는다 | `TestTheRequestedRowCountTravelsBesideTheOneThatArrived` · `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` | yes | yes |
| B4 | 심볼이 빈 행은 행이 되지 않고, 그 읽기는 온전하지도 않다 | `TestAReadingThatLostRowsToBlankSymbolsIsNotAWholeReading`(3중 2개 빈) · `TestAReadingOfNothingButBlankSymbolsDoesNotBecomeThePreviousReading`(전부 빈 → 0행) | yes (2026-07-28: `whole`을 옛 형태로 되돌리면 실패) | yes |

**2026-07-28 정정·해소**. 이전 판은 "두 skip이 같은 조건이라 집합과 행 집합이 어긋나지
않는다는 것은 구조로만 선다"고 적었다. 두 **집합**은 실제로 어긋나지 않았다 — 어긋난 것은
그 읽기가 *온전한가*를 결정하는 비교였고, 그 비교는 걸러지기 **전**의 행 수로 이루어지고
있었다(issues.md I16). `blank_symbol_test.go`가 세 shape(일부 빈·전부 빈·WTS)과 대조군
하나로 이 분기와 `rememberRead` B2를 함께 몬다.
