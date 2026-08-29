# Function Logic Map: `WithBaseURL`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| URL override | Test/explicit endpoint only; never authoritative origin | `official.New` option application | Client remains usable for ordinary reads, but FX authority mint must be disabled |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | client configuration is already sealed | none | option replay is a no-op | replay/race tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimRight` | normalize configured endpoint | no I/O | current source |

## State mutations and fallbacks

- Before sealing, mutates `Client.base` and irreversibly clears authority under `configMu`; after sealing it cannot mutate.

## Safety conclusion

- Safe edit boundary: preserve ordinary test/read behavior while removing authority eligibility.
- High-risk impact: yes — provenance used by risk sizing.
