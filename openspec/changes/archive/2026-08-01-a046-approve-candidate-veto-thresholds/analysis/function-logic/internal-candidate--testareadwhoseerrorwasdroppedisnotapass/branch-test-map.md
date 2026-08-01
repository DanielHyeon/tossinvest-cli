# Branch Test Map: `TestAReadWhoseErrorWasDroppedIsNotAPass`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | simulated read error | this test | existing coverage | no rows plus error |
| B2 | dropped read passes | this test | existing coverage | false |
| B3 | inspect all D3 states | this test | missing accessor during RED | all three |
| B4 | missing reason absent | this test | existing coverage | named reason |
