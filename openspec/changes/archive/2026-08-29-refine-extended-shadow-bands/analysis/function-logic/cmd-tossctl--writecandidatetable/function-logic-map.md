# Function Logic Map: `writeCandidateTable`

- Source: `cmd/tossctl/candidate.go`
- Function: `cmd/tossctl/candidate.go:writeCandidateTable`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: the count rows fold at reportWidth, the values row is new and the collapse alarm prints above the counts. B1-B19 and B21-B23, B28-B34 are base's branches; B20, B25 and B26 became ranges over wrapCounts, and B24 and B27 are new.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| w io.Writer | any | the command | a write error is not consulted; unchanged from base |
| res candidate.CycleResult | one turn | Cycle | only res.Scan.Rejected is read from it |
| report candidateReport | built by buildCandidateReport | the same file | a code missing from ShadowBands is skipped by B23 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the cycle halted | prints the halt reason | no return | existing halt tests |
| B2 | nothing was due | prints the quiet sentence instead of a source count | no return | existing quiet-turn tests |
| B3 | else | prints attempted/responded and the degraded marker | no return | TestTheScanOutputNamesTheMissingSourcesAndTheRetreat |
| B4 | range over missing sources | one line each | no return | same |
| B5 | the source was rate limited | appends the marker | no return | same |
| B6 | some sources were not due | prints the not-due row | no return | existing tests |
| B7 | range over backoff entries | one line each | no return | same |
| B8 | the engine yielded | prints the yield note | no return | existing engine-yield tests |
| B9 | range over readings | one line each | no return | existing tests |
| B10 | the reading was whole | labels it whole rather than short | no return | same |
| B11 | first ranks were held | prints the held sentence | no return | existing held tests |
| B12 | rows were rejected | prints the rejected count | no return | existing rejection tests |
| B13 | range over rejected symbols | one line each | no return | same |
| B14 | range over veto codes | raised/unmeasured per code | no return | existing veto tests |
| B15 | range over veto reasons | one line each | no return | same |
| B16 | range over tally alarms | one line each | no return | existing alarm tests |
| B17 | there are per-source sightings | prints that section | no return | existing sighting tests |
| B18 | range over sightings | one line each | no return | same |
| B19 | range over a sighting's refusals | one line each | no return | same |
| B20 | range over the folded acceleration crossings row | prints one or more lines | no return | TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth; the five-edge case does not fold |
| B21 | range over acceleration not-computed reasons | one line each | no return | existing tests |
| B22 | range over seen_late then extended | one section each | no return | TestTheShadowBandRowsStayInsideEightyColumns |
| B23 | the report has no block for this code | none | continue | a code with no scale is skipped rather than printed empty |
| B24 | NEW - the tally is collapsed | prints the alarm above the counts | no return | TestTheReportRaisesTheAlarmForAScaleThatResolvedNothing |
| B25 | range over the folded crossings row | prints one or more lines | no return | TestTheShadowBandRowsStayInsideEightyColumns; mutation reportWidth 80 -> 200 is RED |
| B26 | NEW - range over the folded values row | prints the distribution | no return | TestTheReportShowsWhereTheValuesWereAndNotOnlyHowManyCrossed |
| B27 | NEW - the quantile base differs from the measured count | prints the discrepancy | no return | unreachable today because every value formatDecimal produces parses; written as a refusal rather than an assumption, and TestTheTallyCarriesTheDistribution pins the equal case |
| B28 | range over a band's unmeasured reasons | one line each | no return | existing tests |
| B29 | switch on the retention outcome | one of three sentences | no return | existing retention tests |
| B30 | the sweep could not finish | prints why | no return | same |
| B31 | the write-ahead log was held | prints that | no return | same |
| B32 | default | prints reclaimed | no return | same |
| B33 | free space was measured | prints the figures | no return | existing space tests |
| B34 | else | prints unmeasured and why | no return | same |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| wrapCounts | NEW - folds a count row at reportWidth with an aligned continuation | total; never drops a part | ast.json calls; candidatebands.go:wrapCounts |
| orderedCountParts / bandCountParts / quantileParts | NEW - the three row bodies | total | ast.json calls |
| sortedCounts | unchanged | total | ast.json calls |

## State mutations and fallbacks

- Writes to w. No state.

## Safety conclusion

- Safe edit boundary: Printing the collapse alarm below the counts. The counts are what a reader takes away and the whole failure was that they looked fine.
- High-risk impact: no - a read-only rendering.
