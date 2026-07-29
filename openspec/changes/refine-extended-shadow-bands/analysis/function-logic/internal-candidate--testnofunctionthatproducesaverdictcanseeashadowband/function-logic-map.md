# Function Logic Map: `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`

- Source: `internal/candidate/band_test.go`
- Function: `internal/candidate/band_test.go:TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Rewritten in this change. It is the guard that stayed green while a band decided a veto.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| the package source at `.` | non-test .go files | parser.ParseDir | a parse error is fatal, never a silent skip |
| verdicts | VetoState, Chase and NEW Verdict | this test | a missing name makes the check pass over that function |
| bandNames | the band types and constructors, plus NEW Crossed and Crossings | this test | same |
| bandFields | NEW - the Verdict's two band slots | this test | same |
| verdictProducers | the floor on how many functions must still be reached (12) | bandguard_test.go | a smaller count is now fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the package did not parse | none | t.Fatalf | a broken package fails here rather than passing |
| B2 | range over parsed packages | none | no return | the whole test |
| B3 | range over files | none | no return | the whole test |
| B4 | range over declarations | none | no return | the whole test |
| B5 | not a function, or returns nothing | none | continue | the whole test |
| B6 | range over the result list | none | no return | the whole test |
| B7 | NEW - range over the names a result type is built from | none | no return | reaches []Verdict (Assess) and would reach *VetoState, []Chase, map[K]Chase |
| B8 | a result names a verdict type | sets produces | no return | the checked count of 12 |
| B9 | the function produces no verdict | none | continue | the whole test |
| B10 | an identifier is in bandNames | t.Errorf | no return | mutation 0.4 - RED on `Crossed`; also RED on the two Measure*Band calls before 0.3 |
| B11 | NEW - a selector reads a Verdict band field outside an assignment's left side | t.Errorf | no return | mutation 0.4 - RED on `v.ExtendedBand` |
| B12 | fewer functions checked than verdictProducers | t.Fatalf | no return | the floor; `checked == 0` would not have caught 12 falling to 1 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| verdictResultNames | reads pointer, slice and map results, not only bare idents | total; nil for a type it does not model | bandguard_test.go |
| assignedTo | separates assembling a band from reading one | total | bandguard_test.go |
| parser.ParseDir / ast.Inspect | the source evidence | a parse error is fatal | ast.json calls |

## State mutations and fallbacks

- None outside the test.

## Safety conclusion

- Safe edit boundary: Widening bandNames to the band field names without the assignment exception. A Verdict has band fields and something has to fill them, so that rule is one no assembly can satisfy.
- High-risk impact: no.
