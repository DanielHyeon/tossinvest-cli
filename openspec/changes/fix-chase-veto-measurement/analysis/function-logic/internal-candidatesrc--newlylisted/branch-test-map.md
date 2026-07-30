# Branch Test Map: `newlyListed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 직전 읽기가 없으면 미상 | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants` · `TestOneSourceServingTwoMarketsDoesNotAnswerAboutTheWrongList` | yes | yes |
| B2 | 직전에 있었으면 `no`, 없었으면 `yes` | `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` · `TestAGenuinelyNewSymbolIsStillFoundAfterAShortReading` · `TestAWTSRowIdentifiedByItsProductCodeIsNotANewEntrantEveryTime` | yes | yes |
