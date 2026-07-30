# Branch Test Map: `Store.FirstRank`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 저장된 위치와 두 자격을 함께 돌려준다 | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` · `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext` · `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan` | yes | yes |
| B2 | 후보가 없으면 found=false | `TestARankOfZeroIsNotAFirstSighting` | — (기존 동작) | yes |
| B3 | Scan 오류 | **커버 없음** (I/O) | no | no |
