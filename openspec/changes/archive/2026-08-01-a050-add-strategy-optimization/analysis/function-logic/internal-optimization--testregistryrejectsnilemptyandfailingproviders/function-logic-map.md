# Function Logic Map: `TestRegistryRejectsNilEmptyAndFailingProviders`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| invalid provider cases | nil, non-setting category, blank owner, empty descriptors, provider error | table fixture | every case must satisfy `errors.Is(ErrInvalidRegistry)` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate all invalid provider cases | test subtests only | continue through matrix | this table-driven test |
| B2 | case does not return typed invalid-registry error | test failure only | fail subtest | this table-driven test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `optimization.BuildRegistry` | exercise fail-closed provider composition | one call per case; no retry | test assertions |

## State mutations and fallbacks

- Test-only local fixtures; registry construction never persists state and no invalid case has a fallback.

## Safety conclusion

- Safe edit boundary: negative registry composition coverage.
- High-risk impact: no direct side effect; protects a high-risk authority boundary.
