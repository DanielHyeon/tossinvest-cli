# Branch Test Map: `TestEntryDecisionIDIsNeverUpdated`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | exact position INSERT token is counted | self | existing | pass |
| B2 | exact UPDATE token is rejected while longer identity token is ignored | self + journal suite | identity substring false positive | pass |
| B3 | scan positive control | self | existing | pass |
| B4 | each parsed UPDATE is checked with exact-token matching | self + journal suite | substring false positive | pass |
| B5 | zero inserters fails the positive control | self | existing | pass |
