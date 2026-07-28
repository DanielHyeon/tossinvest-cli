# Branch Test Map: `officialRanking.Read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 일반 오류는 그대로 래핑 | `internal/candidatesrc` 오류 경로(`TestARateLimitedRankingIsReportedAsOne`의 fake err) | — (기존 동작) | yes |
| B2 | 429는 `candidate.ErrRateLimited`, 그리고 기억은 손대지 않는다 | `TestARateLimitedRankingIsReportedAsOne` · `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` | yes (F1 probe) | yes |
| B3 | 도착한 행마다 요청 수와 3-상태를 싣는다 | `TestTheRequestedRowCountTravelsBesideTheOneThatArrived` · `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` | yes | yes |
| B4 | 심볼이 빈 행은 행이 되지 않는다 | **커버 없음** | no | no |

**정직한 커버리지 기록**: B4(빈 심볼 skip)는 이 change 이전부터 있던 분기이고 지금도 어떤
테스트도 빈 `Symbol`을 가진 `domain.RankingItem`을 넣지 않는다(`rg 'Symbol: *""'` 0건).
이 change가 그 위에 얹은 것은 `rememberRead`의 대칭 skip(B2)뿐이며, **둘이 같은 조건**이라
집합과 행 집합이 어긋나지 않는다는 것은 구조로만 선다. 테스트로는 서 있지 않다.
