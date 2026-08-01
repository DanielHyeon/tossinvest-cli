# Function Logic Map: `ProtectionVerifier.Verify`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed verifier, runtime scope, evidence bytes | verifier has an internal policy source and clock; scope/evidence must exactly match the signed matrix | canonical policy + runtime caller | any parse or authorization error propagates fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | sealed parse fails | reads only verifier-owned artifacts | typed trust/file/signature error | zero verifier and malformed artifact tables |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ProtectionVerifier.parse`, `ProtectionVerifier.verifyParsed` | separate authenticity parsing from current authorization | no fallback or retry | CodeGraph + AST |

## State mutations and fallbacks

- Only verifier-local monotonic policy state is updated; no file, broker, toggle, or engine mutation.

## Safety conclusion

- Safe edit boundary: one sealed entrypoint that cannot accept caller-selected trust or time inputs.
- High-risk impact: yes; dormant authorization prerequisite only.
