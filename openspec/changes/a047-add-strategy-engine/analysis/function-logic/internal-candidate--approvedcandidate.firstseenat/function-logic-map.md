# Function Logic Map: `ApprovedCandidate.FirstSeenAt`

- Source: `internal/candidate/thresholdset.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed receiver | immutable `ApprovedCandidate` value | `AssessApprovedCandidate` mint boundary | zero receiver returns zero scalar |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unconditional value accessor | none | returns a copy of the sealed field | immutable provenance test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | leaf accessor | no error, timeout, retry, or fallback | AST |

## State mutations and fallbacks

- No mutation, allocation with shared backing storage, I/O, or fallback.

## Safety conclusion

- Safe edit boundary: `${name}` exposes only one copied value from the opaque approval.
- High-risk impact: no — branchless leaf, but evidence is retained because the persisted diff intersects the existing function.
