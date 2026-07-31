# Function Logic Map: `forbiddenApprovedBoundaryPackage`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rel` | repository-relative Go package path | import graph built by `auditApprovedCandidateBoundaries` | unknown package is not itself an authority root |
| `forbidden` | complete candidate-isolation authority-root registry | `internal/candidate/isolation_test.go` | any listed root is rejected; sharing this map prevents drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `rel` exists in `forbidden` | none | `true` | `TestApprovedCandidateBoundaryDetectsAllAuthorityRoots` |
| B2 | `rel` is absent | none | `false` | `TestApprovedCandidateBoundaryDetectsAllAuthorityRoots` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| map lookup only | classify the path against the isolation guard's canonical registry | no error/timeout/retry | Go AST + `isolation_test.go` |

## State mutations and fallbacks

- Read-only classification; no state mutation, I/O, fallback, or live binding.

## Safety conclusion

- Safe edit boundary: replace the narrower five-package order map with the existing full authority-root registry; do not add/remove authority roots here.
- High-risk impact: yes — a false negative permits a verdict-derived value to reach order, Guardian, engine, ledger, or domain authority.
