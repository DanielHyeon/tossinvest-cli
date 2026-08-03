# Function Logic Map: `TestTheTracerDrivesEntryRatchetAndExit`

- Source: `internal/app/engine/tracer_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | TestTheTracerDrivesEntryRatchetAndExit | fail closed or test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- Base-revision evidence records the removed or renamed scalar-test path.

## Safety conclusion

- Safe edit boundary: TestTheTracerDrivesEntryRatchetAndExit only.
- High-risk impact: reviewed and regression-tested.
