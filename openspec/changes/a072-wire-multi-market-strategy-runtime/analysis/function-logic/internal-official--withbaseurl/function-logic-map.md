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
| none | returns an option closure | closure replaces base URL | no direct error | `TestAuthorityOriginRejectsConfiguredTransport` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimRight` | normalize configured endpoint | no I/O | current source |

## State mutations and fallbacks

- Mutates only `Client.base`; hardening will also irreversibly mark the client origin overridden.

## Safety conclusion

- Safe edit boundary: preserve ordinary test/read behavior while removing authority eligibility.
- High-risk impact: yes — provenance used by risk sizing.
