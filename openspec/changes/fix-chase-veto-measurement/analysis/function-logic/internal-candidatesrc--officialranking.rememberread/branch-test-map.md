# Branch Test Map: `officialRanking.rememberRead`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 행마다 집합에 넣는다 | `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` | yes | yes |
| B2 | 빈 심볼은 집합에 넣지 않는다 | **커버 없음** | no | no |
| B3 | 첫 호출에서 맵을 만든다 | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants` | yes | yes |
| B4 | 짧은/빈 읽기는 교체하지 않고, 온전한 읽기는 교체한다 | `TestAShortReadingDoesNotBecomeTheYardstickForTheNextWholeOne` · `TestAGenuinelyNewSymbolIsStillFoundAfterAShortReading` · `TestAnEmptyReadingDoesNotBecomeThePreviousReading` | yes (F2가 무조건 swap에서 실행 확인) | yes |
| B5 | 장애 뒤·시계 역행 뒤의 기억은 답하지 않는다 | `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` · `TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore` · `TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh` | yes (F1 probe) | yes |

**정직한 커버리지 기록**: B2(빈 심볼)는 커버되지 않는다. `Read`의 같은 skip과 짝이어야
한다는 것이 이 분기의 존재 이유인데, 두 skip이 함께 움직인다는 것을 잡는 테스트는 없다.
`officialRanking.Read` B4와 같은 공백이다.
