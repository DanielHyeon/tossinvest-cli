# Function Logic Map: `failingProvider.OwnerChange`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fixture owner | test-provided string | `failingProvider.owner` | returned unchanged so registry can attribute provider error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless accessor happy path | none | fixture owner | `TestRegistryRejectsNilEmptyAndFailingProviders/provider_error` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | test-only value accessor | no error/retry | AST and provider-error subtest |

## State mutations and fallbacks

- No mutation, I/O, fallback, or production binding.

## Safety conclusion

- Safe edit boundary: test fixture only.
- High-risk impact: no.
