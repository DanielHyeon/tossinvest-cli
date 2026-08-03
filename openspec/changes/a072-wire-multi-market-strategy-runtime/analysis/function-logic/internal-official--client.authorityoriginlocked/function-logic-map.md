# Function Logic Map: `Client.authorityOriginLocked`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| caller-held config read/write lock | sealed default base and exact constructor transport | `official.New` | false on any incomplete/configured state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy predicate evaluation | none | true only for complete immutable production origin | authority boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hcTransport` | safely require the concrete private transport type | caller holds `configMu`; no I/O | AST |

## State mutations and fallbacks

- Does not acquire the lock itself so an authoritative read can retain one lock across predicate and network request.

## Safety conclusion

- Safe edit boundary: configuration seal, endpoint and concrete transport pointer identity are conjunctive.
- High-risk impact: yes, this is the origin authority predicate.
