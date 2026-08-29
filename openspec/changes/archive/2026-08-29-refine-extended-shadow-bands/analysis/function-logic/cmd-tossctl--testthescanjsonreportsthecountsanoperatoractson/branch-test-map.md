# Branch Test Map: `TestTheScanJSONReportsTheCountsAnOperatorActsOn`

- Source: `cmd/tossctl/candidate_test.go`
- Function: `cmd/tossctl/candidate_test.go:TestTheScanJSONReportsTheCountsAnOperatorActsOn`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the command failed | the command must succeed for the rest to mean anything | no | yes |
| B2 | the output is not JSON | RED observed: with `crossed` still decoded as map[string]int, `json: cannot unmarshal array into ...` | yes | yes |
| B3 | the market is wrong | existing | no | yes |
| B4 | the veto tally is wrong | existing | no | yes |
| B5 | a pass was counted | existing - an absent threshold is not a pass | no | yes |
| B6 | the passed note is missing | existing | no | yes |
| B7 | THRESHOLD_ABSENT was not counted | existing | no | yes |
| B8 | the acceleration keys are incomplete | existing | no | yes |
| B9 | range over seen_late and extended | existing | no | yes |
| B10 | a code has no block | existing | no | yes |
| B11 | the block's total is wrong | existing | no | yes |
| B12 | NEW - the crossed list is not the length of the scale | a truncated list must not read as a scale nobody crossed | no | yes |
| B13 | NEW - range over the crossed entries | the order assertion | no | yes |
| B14 | NEW - an entry is out of scale order | the map form emitted "0","10","100","2", ... which is what this refuses | no | yes |
