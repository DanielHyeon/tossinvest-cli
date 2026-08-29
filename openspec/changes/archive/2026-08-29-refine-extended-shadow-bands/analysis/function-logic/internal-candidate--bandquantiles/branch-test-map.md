# Branch Test Map: `bandQuantiles`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:bandQuantiles`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no values | TestAnUnmeasuredTallyHasNoQuantiles | yes | yes |
| B2 | range over BandQuantilePoints | TestTheTallyCarriesTheDistributionAndNotOnlyTheCrossings | yes | yes |
| B3 | the computed rank is below 1 | same test, the min position | no | yes |
| B4 | the computed rank is past the end | same test, the max position | no | yes |
