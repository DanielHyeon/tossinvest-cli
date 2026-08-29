# Branch Test Map: `shadowBandsFor`

- Source: `internal/candidate/watch.go`
- Function: `internal/candidate/watch.go:shadowBandsFor`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | both records are measured and reach the verdict | TestAVetoWithNoThresholdStillLeavesAShadowRecord, TestABandNamesTheSameMissingInputTheVetoWould, TestNoFunctionThatProducesAVerdictCanSeeAShadowBand | no | yes |
