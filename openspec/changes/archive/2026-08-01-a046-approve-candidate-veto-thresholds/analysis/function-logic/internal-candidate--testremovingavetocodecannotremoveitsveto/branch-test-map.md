# Branch Test Map: `TestRemovingAVetoCodeCannotRemoveItsVeto`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | wrong cardinality | this test | missing accessor during RED | three |
| B2 | compare order | this test | exported array writable | stable copy |
| B3 | wrong code | this test | exported array writable | exact D3 order |
| B4 | copy mutation removes pass check | this test | old exported variable mutable | false |
| B5 | copy mutation removes veto | this test | old exported variable mutable | true |
| B6 | copy mutation changes tally | this test | old exported variable mutable | raised count retained |
