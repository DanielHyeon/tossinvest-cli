# Function Logic Map: `Chase.Raised`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| three veto states | raised, clear, unmeasured | private D3 order | append dangerous codes only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied order | local slice only | ordered result | veto unit suite |
| B2 | state dangerous | append local code | continue | raised-veto tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `Chase.State`, `Dangerous` | ordered raised list | pure | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only a fresh result slice; no shared state or fallback.

## Safety conclusion

- Safe edit boundary: immutable ordering access.
- High-risk impact: no.
