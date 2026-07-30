# Branch Test Map: `officialRanking.rememberRead`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 행마다 집합에 넣는다 | `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` | yes | yes |
| B2 | 빈 심볼은 집합에 넣지 않고, 그래서 그 읽기는 온전하지 않다 | `TestAReadingThatLostRowsToBlankSymbolsIsNotAWholeReading` · `TestAReadingOfNothingButBlankSymbolsDoesNotBecomeThePreviousReading` · `TestAGenuinelyNewSymbolIsStillFoundAfterABlankReading`(대조군) | yes (2026-07-28: `whole`을 옛 형태로 되돌리면 3건 실패) | yes |
| B3 | 첫 호출에서 맵을 만든다 | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants` | yes | yes |
| B4 | 짧은/빈 읽기는 교체하지 않고, 온전한 읽기는 교체한다 | `TestAShortReadingDoesNotBecomeTheYardstickForTheNextWholeOne` · `TestAGenuinelyNewSymbolIsStillFoundAfterAShortReading` · `TestAnEmptyReadingDoesNotBecomeThePreviousReading` | yes (F2가 무조건 swap에서 실행 확인) | yes |
| B5 | 장애 뒤·시계 역행 뒤의 기억은 답하지 않는다 | `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` · `TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore` · `TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh` | yes (F1 probe) | yes |

**2026-07-28 해소**: B2는 이제 커버된다. 이전 판은 "두 skip이 함께 움직인다는 것을 잡는
테스트가 없다"고 적었는데, 실제 결함은 두 skip이 아니라 **온전 판정**이 그 skip을 보지
않는다는 것이었다(issues.md I16). `blank_symbol_test.go`가 그 자리를 몬다.
