# Function Logic Map: `verifyRunLockPath`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| evidence record path | default/profile or explicit override record | verify command | returns a marker beside that exact record |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| tail | always | none | sibling `verify-run.lock` path | existing runlock path tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.Dir/Join` | retain record-directory isolation | pure | base AST |

## State mutations and fallbacks

- Pure path derivation; this pre-existing helper remains unchanged.

## Safety conclusion

- Safe edit boundary: no implementation edit; base evidence is included because adjacent insertion touches the diff boundary.
- High-risk impact: no.
