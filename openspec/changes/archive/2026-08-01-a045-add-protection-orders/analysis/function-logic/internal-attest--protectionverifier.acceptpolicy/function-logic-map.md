# Function Logic Map: `ProtectionVerifier.acceptPolicy`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| canonical policy generation/digest | generation positive; digest canonical and already validated | policy loader | rollback/reuse returns `ErrProtectionTrust` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | lower generation or equal generation with changed digest | mutex only | `ErrProtectionTrust` | direct-call rollback/reuse |
| B2 | first observation or higher generation | records generation/digest | nil | initial verify and rotation |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sync.Mutex` | serialize monotonic state across concurrent verifies | lock released on all paths | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only in-memory highest accepted generation/digest; no rollback fallback.

## Safety conclusion

- Safe edit boundary: fail-closed anti-rollback latch scoped to one sealed verifier instance.
- High-risk impact: yes; prevents restored old trust files from regaining authority.
