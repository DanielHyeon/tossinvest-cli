# Branch Test Map: `TestTheZeroChaseDoesNotPassTheVeto`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero incorrectly passes | this test | existing guard | false |
| B2 | zero not marked missing | this test | existing guard | true missing |
| B3 | fewer than three missing codes | this test | missing accessor during RED | exactly three |
| B4 | zero incorrectly vetoed | this test | existing guard | false |
