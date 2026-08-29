# Function Logic Map: `firstLegRequestDigest`

- Source: `internal/journal/strategy_first_leg_atomic.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| prepared request | exact q_final/strategy/campaign/router plus optional weekly key | preparation | marshal/hash error aborts |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | JSON marshal fails | none | error | structural unit coverage |
| success | canonical preimage marshals | none | SHA-256 | replay tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | canonical closed preimage | returns error | CodeGraph + AST |

## State mutations and fallbacks

- Pure digest; optional weekly binding changes replay identity and cannot be dropped.

## Safety conclusion

- Safe edit boundary: extend the closed preimage only; retain versioned SHA-256 semantics.
- High-risk impact: yes/no
