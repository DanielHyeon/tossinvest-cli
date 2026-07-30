# Branch Test Map: `orderedCounts`

- Source: `cmd/tossctl/candidate.go`
- Function: `cmd/tossctl/candidate.go:orderedCounts`

RED/GREEN is what was actually observed while implementing this change. `no` in the RED
column means the branch is base behaviour this change did not alter and no failing state was
manufactured for it; the test named is the one that covers it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | there are no keys | existing scan report tests; a code with no scale renders "none" | no | yes |
