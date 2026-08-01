# Function Logic Map: `hashCapability`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability token | nonempty opaque string from caller | apply/preview contract | caller validation handles blank token |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | deterministic hash | no mutation | SHA-256 hex result | lifecycle capability tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SHA-256 | stable lookup key | no I/O or retry | AST |

## State mutations and fallbacks

- Pure token transformation; persisted storage never contains the raw capability.

## Safety conclusion

- Safe edit boundary: pure hash helper.
- High-risk impact: no.
