# Branch Test Map: `buildCandidateReport`

- Source: `cmd/tossctl/candidate.go`
- Function: `cmd/tossctl/candidate.go:buildCandidateReport`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range over NotDue source ids | existing scan report tests | no | yes |
| B2 | range over missing sources | TestTheScanOutputNamesTheMissingSourcesAndTheRetreat | no | yes |
| B3 | range over backoff entries | same | no | yes |
| B4 | range over readings in sorted source order | existing scan report tests | no | yes |
| B5 | range over per-source sightings | existing first-sighting tests | no | yes |
| B6 | range over one source's unmeasured reasons | same | no | yes |
| B7 | the passed count is not zero | existing passed-note tests | no | yes |
| B8 | range over raised veto codes | TestTheScanJSONReportsTheCountsAnOperatorActsOn | no | yes |
| B9 | range over unmeasured veto codes | same | no | yes |
| B10 | range over veto reasons | same | no | yes |
| B11 | range over acceleration not-computed reasons | same | no | yes |
| B12 | range over the shadow band tallies | same | no | yes |
| B13 | range over one tally's unmeasured reasons | same | no | yes |
| B14 | NEW - range over BandsFor(code) | TestTheScanJSONReportsTheCountsAnOperatorActsOn, order assertion | yes | yes |
| B15 | NEW - range over the tally's quantiles | TestTheReportShowsWhereTheValuesWereAndNotOnlyHowManyCrossed | yes | yes |
