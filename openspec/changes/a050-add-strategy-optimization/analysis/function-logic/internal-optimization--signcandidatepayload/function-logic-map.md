# Function Logic Map: `signCandidatePayload`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque capability and typed candidate payload | payload has deterministic struct field order | Preview/Apply payload values | marshal error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | JSON marshal fails or succeeds | no persistent mutation | error or HMAC-SHA256 hex | payload MAC test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal`, `hmac.New` | canonical bytes and signature | marshal error propagates | AST |

## State mutations and fallbacks

- Pure helper; the raw capability is never inserted in SQLite.

## Safety conclusion

- Safe edit boundary: candidate tamper evidence.
- High-risk impact: no LIVE authority.
