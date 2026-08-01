# Function Logic Map: `ProtectionVerifier.parse`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| verifier internals | non-empty canonical source, verifier-owned clock, platform owner metadata | provisioned sealed verifier | `ErrProtectionTrust`/`ErrProtectionFile` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | verifier/source/clock absent | none | `ErrProtectionTrust` | zero verifier |
| B2 | current owner metadata unavailable | none | `ErrProtectionFile` | non-Unix helper compiles fail-closed |
| B3 | canonical policy load fails | read-only artifact access | propagated typed error | policy permission/path/canonical JSON tests |
| B4 | signed envelope/root parse fails | bounded read and cryptographic verification only | typed file/trust/signature error | signed adversarial tables |
| B5 | fully validated policy generation/digest is rejected | monotonic state lock only | `ErrProtectionTrust` | direct-call rollback/reuse test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| policy loader + signed parser | pin paths/owner/digest and authenticate envelope | no fallback; every error terminates | CodeGraph + AST |

## State mutations and fallbacks

- Records the highest accepted policy generation/digest under a mutex; does not mutate artifacts.

## Safety conclusion

- Safe edit boundary: production cannot inject source path, UID, digest, key, or clock.
- High-risk impact: yes; dormant trust boundary.
