# Function Logic Map: `CollapsedAlarm`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:BandTally.CollapsedAlarm`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

New in this change. The sentence Collapsed() is reported with.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| t BandTally (receiver) | any tally | TallyBands | empty string when the scale resolved something |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the tally is not collapsed | none | empty string, so the caller renders nothing | TestATallyThatResolvedNothingSaysSo, second case |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| t.Collapsed | the judgement, which is not duplicated here | total | ast.json calls |
| fmt.Sprintf | carries the numbers that produced the alarm | total | ast.json calls |

## State mutations and fallbacks

- None.

## Safety conclusion

- Safe edit boundary: Wording the judgement here as well as in Collapsed(). Two screens must not be able to disagree about whether a scale resolved anything - the rule tallyAlarm already follows.
- High-risk impact: no.
