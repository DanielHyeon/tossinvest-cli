# Function Logic Map: `ProtectionVerifier.verifyParsed`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| parsed value, verifier, runtime scope/evidence | parsed value belongs to this verifier; final policy/root/key are current | signed parse + canonical policy | stale, foreign, revoked, mismatched, or expired authority fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | verifier identity/clock absent or parsed by another verifier | none | `ErrProtectionTrust` | sealed verifier identity test |
| B2 | exact scope or evidence prevalidation fails | caller-controlled hashing only | typed scope/evidence error | scope/evidence tables |
| B3 | deterministic test hook is present after evidence | test-only callback effect | continue to final reload | revoke-after-evidence race test |
| B4 | final policy reload fails | read-only artifact access | propagated error | policy mutation/permission tests |
| B5 | current policy predates parse or reuses generation with changed digest | none | `ErrProtectionTrust` | parse-then-policy rollback/reuse |
| B6 | current root reload/digest/generation validation fails | read-only artifact access | propagated error | root canonical and digest tables |
| B7 | current key ID absent | none | `ErrProtectionTrust` | unknown/replaced key test |
| B8 | signature invalid under current key | none | `ErrProtectionSignature` | envelope tamper/signer replacement |
| B9 | final verifier time is before matrix issue | clock read only | `ErrProtectionInvalid` | future issue table |
| B10 | final verifier time is at/after matrix expiry | clock read only | `ErrProtectionExpired` | expiry table |
| B11 | current key status/window rejects | none | `ErrProtectionTrust` | revoke-after-evidence and key-window tables |
| B12 | fully validated policy violates verifier monotonic latch | mutex-protected observation only | `ErrProtectionTrust` | cross-call rollback/reuse |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| matrix prevalidator, current policy/root loader, signature verifier | complete unbounded evidence work first, then enforce current trust and time immediately before authority construction | no cache/fallback/retry | CodeGraph + AST |

## State mutations and fallbacks

- Updates only verifier-local monotonic observation state; returns exact matched authority.

## Safety conclusion

- Safe edit boundary: final authorization revalidates the current root and signer after parsing.
- High-risk impact: yes; security authorization boundary, still not engine-wired.
