# Function Logic Map: `ProtectionTrustRoot.validate`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| untrusted root | positive monotone generation, 1-16 strictly key-ID-sorted unique valid keys | canonical root JSON | `ErrProtectionTrust` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | version/generation/count; per-key validation; strict ordering; duplicate ID/public key | none | trust error | malformed, reordered and lifecycle tables |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| key validation | role/algorithm/key bytes/window/revocation contract | no fallback | CodeGraph + AST |

## State mutations and fallbacks

- None.

## Safety conclusion

- Safe edit boundary: add generation and semantic ordering without accepting alternate roots.
- High-risk impact: yes.
