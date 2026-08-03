# Function Logic Map: `Client.AuthorityOrigin`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| client provenance state | non-nil, default base, eligible, exact constructor transport | official.New | returns zero token and false on any mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil receiver | none | zero token, false | configured-origin test matrix |
| B2 | sealed origin predicate fails under read lock | none | zero token, false | authority and replay tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `authorityOriginLocked` | evaluate the complete sealed provenance predicate | pure under `configMu.RLock` | AST |

## State mutations and fallbacks

- Issues the opaque origin token only when all constructor-owned provenance facts still match.

## Safety conclusion

- Safe edit boundary: every explicit endpoint or client override permanently clears eligibility.
- High-risk impact: yes, prevents arbitrary HTTP responses minting FX authority.
