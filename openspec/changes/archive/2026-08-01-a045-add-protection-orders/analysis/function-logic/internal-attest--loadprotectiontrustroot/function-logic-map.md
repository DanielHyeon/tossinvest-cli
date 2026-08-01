# Function Logic Map: `loadProtectionTrustRoot`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current canonical policy | exact root path/owner/digest/generation | sealed policy loader | mismatch or unsafe filesystem fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | bounded artifact read; digest compare; canonical JSON/keyset validation | read-only | `ErrProtectionTrust` | root replacement, generation and canonical tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readProtectionArtifact` | race-resistant bounded load | no fallback | CodeGraph + AST |

## State mutations and fallbacks

- None. Root must be reloaded at final authorization.

## Safety conclusion

- Safe edit boundary: bind every load to the currently re-read sealed policy.
- High-risk impact: yes.
