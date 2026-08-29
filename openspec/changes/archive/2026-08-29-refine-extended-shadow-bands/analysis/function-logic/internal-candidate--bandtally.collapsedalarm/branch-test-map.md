# Branch Test Map: `CollapsedAlarm`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:BandTally.CollapsedAlarm`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the tally is not collapsed | TestATallyThatResolvedNothingSaysSo, second case | yes | yes |
