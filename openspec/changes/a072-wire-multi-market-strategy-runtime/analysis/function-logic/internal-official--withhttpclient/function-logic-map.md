# Function Logic Map: `WithHTTPClient`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| HTTP client override | Any explicit override | `official.New` option application | Ordinary client works; authority eligibility must be irreversibly false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | returns an option closure | closure replaces transport and marks override | no direct error | `TestAuthorityOriginRejectsConfiguredTransport` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | field assignment only | no I/O | current source |

## State mutations and fallbacks

- Mutates `Client.hc`; hardening adds provenance state and does not alter request behavior.

## Safety conclusion

- Safe edit boundary: transport override remains supported for non-authoritative reads/tests.
- High-risk impact: yes — arbitrary transport must not mint official FX authority.
