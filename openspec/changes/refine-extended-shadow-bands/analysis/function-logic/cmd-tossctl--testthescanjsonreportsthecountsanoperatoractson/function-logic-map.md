# Function Logic Map: `TestTheScanJSONReportsTheCountsAnOperatorActsOn`

- Source: `cmd/tossctl/candidate_test.go`
- Function: `cmd/tossctl/candidate_test.go:TestTheScanJSONReportsTheCountsAnOperatorActsOn`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: the decoded shape of `crossed` went from a map to a list and the scale's order is now asserted. B1-B11 are base's; B12-B14 are new.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| a one-candidate fixture store | a fake clock and one discovery row | withCandidateFixture | no network, no real store |
| the scan's JSON output | must decode | runCandidate | a decode failure is fatal - that is how the map-to-list change was caught |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the command failed | none | t.Fatalf | the command must succeed for the rest to mean anything |
| B2 | the output is not JSON | none | t.Fatalf | RED observed: with `crossed` still decoded as map[string]int, `json: cannot unmarshal array into ...` |
| B3 | the market is wrong | none | t.Errorf | existing |
| B4 | the veto tally is wrong | none | t.Errorf | existing |
| B5 | a pass was counted | none | t.Errorf | existing - an absent threshold is not a pass |
| B6 | the passed note is missing | none | t.Error | existing |
| B7 | THRESHOLD_ABSENT was not counted | none | t.Errorf | existing |
| B8 | the acceleration keys are incomplete | none | t.Errorf | existing |
| B9 | range over seen_late and extended | none | no return | existing |
| B10 | a code has no block | none | t.Fatalf | existing |
| B11 | the block's total is wrong | none | t.Errorf | existing |
| B12 | NEW - the crossed list is not the length of the scale | none | t.Fatalf | a truncated list must not read as a scale nobody crossed |
| B13 | NEW - range over the crossed entries | none | no return | the order assertion |
| B14 | NEW - an entry is out of scale order | none | t.Errorf | the map form emitted "0","10","100","2", ... which is what this refuses |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| candidate.BandsFor | NEW - the expected order, taken from the source of truth rather than restated | total | ast.json calls |
| runCandidate / withCandidateFixture / json.Unmarshal | unchanged | total | ast.json calls |

## State mutations and fallbacks

- None outside the test's temp directory.

## Safety conclusion

- Safe edit boundary: Restating the expected order as a literal. The point of the assertion is that the JSON follows the scale, not that it follows a copy of it.
- High-risk impact: no.
