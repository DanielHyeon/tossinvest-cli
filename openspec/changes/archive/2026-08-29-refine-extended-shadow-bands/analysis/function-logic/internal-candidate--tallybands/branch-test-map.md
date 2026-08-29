# Branch Test Map: `TallyBands`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:TallyBands`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range over BandsFor(code) seeding the Crossed map | TestTheTallyHasOneColumnPerBandAndCountsThemRight | no | yes |
| B2 | range over the input records | TestTheBandTallyAccountsForEveryCandidate | no | yes |
| B3 | the record is unmeasured or belongs to another code | TestABandOfAnotherCodeIsNotCountedAsMeasured, TestAnUnmeasuredTallyHasNoQuantiles | no | yes |
| B4 | a measured record of the wrong code, inside B3 | TestABandOfAnotherCodeIsNotCountedAsMeasured | no | yes |
| B5 | NEW - the record's rendered Value parses as a rational | TestTheTallyCarriesTheDistributionAndNotOnlyTheCrossings | yes (the quantile tests failed before this line existed) | yes |
| B6 | range clearing the per-record `seen` set | TestTheBandTallyAccountsForEveryCandidate | no | yes |
| B7 | range over the record's crossings | TestTheTallyHasOneColumnPerBandAndCountsThemRight | no | yes |
| B8 | the crossing names a band this code does not shadow, or a duplicate | TestABandOfAnotherCodeIsNotCountedAsMeasured | no | yes |
| B9 | the crossing is true | TestTheTallyHasOneColumnPerBandAndCountsThemRight | no | yes |
