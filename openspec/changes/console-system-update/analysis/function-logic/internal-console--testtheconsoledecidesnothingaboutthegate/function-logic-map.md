# Function Logic Map: `TestTheConsoleDecidesNothingAboutTheGate`

- Source: `internal/console/engineproc_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| package sources | all console Go files | static scanner | banned interlock decision term fails |
| exemptions | exact settings files allowed to name config block only | closed map | no prefix/suffix expansion |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | iterate files and extend banned set outside exact exemptions | none | test failure |
| B3-B4 | scan banned terms and report any match | none | test failure |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles`, `nonCommentLines` | inspect production source only | deterministic | AST |
| `strings.Contains` | enforce closed vocabulary boundary | exact text | AST |

## State mutations and fallbacks

- This change only applies gofmt map alignment; guard semantics are unchanged.

## Safety conclusion

- Safe edit boundary: mechanical formatting only.
- High-risk impact: yes — console must not reimplement engine interlock decisions.
