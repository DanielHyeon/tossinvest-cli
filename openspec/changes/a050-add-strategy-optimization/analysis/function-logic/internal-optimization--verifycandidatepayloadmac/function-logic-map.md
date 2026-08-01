# Function Logic Map: `verifyCandidatePayloadMAC`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque capability, typed payload, persisted MAC | HMAC must match exactly in constant time | candidate DB row | false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | signing/encoded MAC decode fails | none | false | payload tamper test |
| B2 | constant-time comparison | none | match result | payload tamper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `signCandidatePayload`, `hmac.Equal` | recompute and compare MAC | errors return false | AST |

## State mutations and fallbacks

- Pure verification; no raw capability is persisted.

## Safety conclusion

- Safe edit boundary: candidate tamper evidence.
- High-risk impact: no LIVE authority.
