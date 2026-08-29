# Branch Test Map: `Collapsed`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:BandTally.Collapsed`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nothing was measured (first disjunct) / the code has no scale at all (second disjunct, added in §5.3 for R3) | TestATallyThatResolvedNothingSaysSo third case; TestATallyWithNoScaleIsNotReportedAsCollapsed | yes for both — the second was RED before the disjunct existed: two measured near_high records reported themselves collapsed and carried an alarm sentence | yes |
| B2 | range over the per-band counts | TestATallyThatResolvedNothingSaysSo | no | yes |
| B3 | a count is neither 0 nor Measured | TestATallyThatResolvedNothingSaysSo, second case | yes | yes |
