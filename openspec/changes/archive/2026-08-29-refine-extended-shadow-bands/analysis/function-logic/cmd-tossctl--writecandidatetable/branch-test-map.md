# Branch Test Map: `writeCandidateTable`

- Source: `cmd/tossctl/candidate.go`
- Function: `cmd/tossctl/candidate.go:writeCandidateTable`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the cycle halted | existing halt tests | no | yes |
| B2 | nothing was due | existing quiet-turn tests | no | yes |
| B3 | else | TestTheScanOutputNamesTheMissingSourcesAndTheRetreat | no | yes |
| B4 | range over missing sources | same | no | yes |
| B5 | the source was rate limited | same | no | yes |
| B6 | some sources were not due | existing tests | no | yes |
| B7 | range over backoff entries | same | no | yes |
| B8 | the engine yielded | existing engine-yield tests | no | yes |
| B9 | range over readings | existing tests | no | yes |
| B10 | the reading was whole | same | no | yes |
| B11 | first ranks were held | existing held tests | no | yes |
| B12 | rows were rejected | existing rejection tests | no | yes |
| B13 | range over rejected symbols | same | no | yes |
| B14 | range over veto codes | existing veto tests | no | yes |
| B15 | range over veto reasons | same | no | yes |
| B16 | range over tally alarms | existing alarm tests | no | yes |
| B17 | there are per-source sightings | existing sighting tests | no | yes |
| B18 | range over sightings | same | no | yes |
| B19 | range over a sighting's refusals | same | no | yes |
| B20 | range over the folded acceleration crossings row | TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth; the five-edge case does not fold | no | yes |
| B21 | range over acceleration not-computed reasons | existing tests | no | yes |
| B22 | range over seen_late then extended | TestTheShadowBandRowsStayInsideEightyColumns | no | yes |
| B23 | the report has no block for this code | a code with no scale is skipped rather than printed empty | no | yes |
| B24 | NEW - the tally is collapsed | TestTheReportRaisesTheAlarmForAScaleThatResolvedNothing | yes | yes |
| B25 | range over the folded crossings row | TestTheShadowBandRowsStayInsideEightyColumns; mutation reportWidth 80 -> 200 is RED | yes | yes |
| B26 | NEW - range over the folded values row | TestTheReportShowsWhereTheValuesWereAndNotOnlyHowManyCrossed | yes | yes |
| B27 | NEW - the quantile base differs from the measured count | unreachable today because every value formatDecimal produces parses; written as a refusal rather than an assumption, and TestTheTallyCarriesTheDistribution pins the equal case | no - unreachable today, see the branch row | n/a |
| B28 | range over a band's unmeasured reasons | existing tests | no | yes |
| B29 | switch on the retention outcome | existing retention tests | no | yes |
| B30 | the sweep could not finish | same | no | yes |
| B31 | the write-ahead log was held | same | no | yes |
| B32 | default | same | no | yes |
| B33 | free space was measured | existing space tests | no | yes |
| B34 | else | same | no | yes |
