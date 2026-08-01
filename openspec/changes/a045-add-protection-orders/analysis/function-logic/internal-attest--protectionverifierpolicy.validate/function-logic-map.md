# Function Logic Map: `protectionVerifierPolicy.validate`

- Source: `internal/attest/protection_signature.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| policy + canonical source path | v1, positive generation, three absolute canonical paths in separate parents, pinned SHA-256 | root-owned policy artifact | `ErrProtectionTrust` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unsupported version or zero generation | none | trust error | malformed policy table |
| B2 | attestation path empty/relative/non-clean/wrong basename | none | trust error | path table |
| B3 | trust-root path empty/relative/non-clean/wrong basename | none | trust error | path table |
| B4 | digest is not canonical SHA-256 | none | trust error | digest table |
| B5 | iterate policy, attestation and root paths | local map only | continue | separated layout happy path |
| B6 | any two artifacts share a parent | none | trust error | same-parent test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| path canonicalization and digest validator | prevent aliasing and unpinned root selection | no fallback | CodeGraph + AST |

## State mutations and fallbacks

- Local duplicate-parent map only; no filesystem mutation.

## Safety conclusion

- Safe edit boundary: canonical policy is data-only and cannot name alternate basenames or colocate mutable runtime data.
- High-risk impact: yes; trust-source definition.
