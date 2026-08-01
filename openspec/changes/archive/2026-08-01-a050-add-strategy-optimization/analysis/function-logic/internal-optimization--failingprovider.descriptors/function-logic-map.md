# Function Logic Map: `failingProvider.Descriptors`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | any test context | caller | deliberately returns no descriptors and a deterministic source error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless failure fixture happy path | none | nil descriptors plus source-unavailable error | `TestRegistryRejectsNilEmptyAndFailingProviders/provider_error` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Errorf` | creates deterministic fixture error | no retry | AST and provider-error subtest |

## State mutations and fallbacks

- No mutation or production binding; this exists only to exercise provider failure propagation.

## Safety conclusion

- Safe edit boundary: test fixture only.
- High-risk impact: no.
