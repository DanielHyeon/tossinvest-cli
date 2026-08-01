# Function Logic Map: `RiskGuardian.PolicyVersion`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Guardian | constructed with non-empty policy version | constructor | returns frozen version |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional); accessor called | none | version string | direct strategy snapshot mismatch/success tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | immutable field accessor | no error path | AST |

## State mutations and fallbacks

- Pure accessor; no mutation.

## Safety conclusion

- Safe edit boundary: read-only policy identity.
- High-risk impact: no.
