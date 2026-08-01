# Function Logic Map: `parseSignedProtectionCapability`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed verifier + current policy | verifier-owned policy source and runtime owner | immutable verifier state | any absent/unsafe/currently unauthorized input fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | read attestation/policy/root; decode canonical envelope/payload; validate matrix; resolve key; verify signature | bounded read only | typed file/trust/signature/claim error | signed adversarial and parse-before-revoke tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| artifact/policy/root loaders | establish filesystem and trust boundary | no retry, fallback or TOFU | CodeGraph + AST |

## State mutations and fallbacks

- Read-only. The returned internal parsed value is non-authoritative and retains signed bytes/key ID, not a copied authorization key.

## Safety conclusion

- Safe edit boundary: move all raw policy inputs behind a sealed verifier and force current-root revalidation later.
- High-risk impact: yes; dormant authorization prerequisite only.
