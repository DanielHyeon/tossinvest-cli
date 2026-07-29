# Branch Test Map: `assessOne`

- Source: `internal/candidate/watch.go`
- Function: `internal/candidate/watch.go:assessOne`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range over rows, grouping observations by source | existing acceleration coverage through Assess | no | yes |
| B2 | first sight of a source id inside B1 | same | no | yes |
| B3 | range over the sorted source ids | same | no | yes |
| band assembly | the two shadow records still reach the verdict after the statements moved out | TestAVetoWithNoThresholdStillLeavesAShadowRecord, TestTheBandTallyAccountsForEveryCandidate | no | yes |
| the back door | `if v.ExtendedBand.Crossed("6") { v.Chase.Extended = RaisedVeto() }` inserted here | TestNoFunctionThatProducesAVerdictCanSeeAShadowBand | yes (2026-07-29, two messages) | yes after revert |
