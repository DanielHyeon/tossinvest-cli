# Function Logic Map: `TestProductionEngineAssemblesPairedUnwiredReadinessProvider`

- Source: `internal/app/engine/a071_readiness_assembly_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| isolated engine config and official stub | production defaults, no manifest pin | engine test harness | fail startup/assertion; never open automation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | engine startup succeeds | local test DB/config creation | fail test on error | named test |
| B2 | readiness provider is absent | none | fail test | named test |
| B3 | default readiness opens entry | none | fail test | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `startEngine` | exercise real production assembly with stubbed official host | fail test on startup error | CodeGraph + AST |

## State mutations and fallbacks

- Test-only config/database mutation inside an isolated temporary directory.

## Safety conclusion

- Safe edit boundary: production assembly assertions only
- High-risk impact: no (test); protects default OFF behavior
