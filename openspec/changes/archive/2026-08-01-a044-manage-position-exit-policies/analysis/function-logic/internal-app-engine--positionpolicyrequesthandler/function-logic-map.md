# Function Logic Map: `positionPolicyRequestHandler`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| HTTP request | JSON POST, bounded single document | loopback client | reject method/content/body/unknown/trailing |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | method, media type, bounded body, framing, decode, or command failure | no direct state mutation | typed response | transport tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| command callback | invoke narrow preview/apply DTO | context timeout from server | AST |

## State mutations and fallbacks

- Decoder is strict and request DTO determines whether scope fields are permitted.

## Safety conclusion

- Safe edit boundary: split Preview Request from token-only ApplyRequest so unknown mutation scope is rejected.
- High-risk impact: yes
