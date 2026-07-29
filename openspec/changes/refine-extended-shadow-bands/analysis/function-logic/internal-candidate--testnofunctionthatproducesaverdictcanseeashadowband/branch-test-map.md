# Branch Test Map: `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`

- Source: `internal/candidate/band_test.go`
- Function: `internal/candidate/band_test.go:TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the package did not parse | a broken package fails here rather than passing | no | yes |
| B2 | range over parsed packages | the whole test | no | yes |
| B3 | range over files | the whole test | no | yes |
| B4 | range over declarations | the whole test | no | yes |
| B5 | not a function, or returns nothing | the whole test | no | yes |
| B6 | range over the result list | the whole test | no | yes |
| B7 | NEW - range over the names a result type is built from | reaches []Verdict (Assess) and would reach *VetoState, []Chase, map[K]Chase | yes | yes |
| B8 | a result names a verdict type | the checked count of 12 | no | yes |
| B9 | the function produces no verdict | the whole test | no | yes |
| B10 | an identifier is in bandNames | mutation 0.4 - RED on `Crossed`; also RED on the two Measure*Band calls before 0.3 | yes | yes |
| B11 | NEW - a selector reads a Verdict band field outside an assignment's left side | mutation 0.4 - RED on `v.ExtendedBand` | yes | yes |
| B12 | fewer functions checked than verdictProducers | the floor; `checked == 0` would not have caught 12 falling to 1 | no | yes |
