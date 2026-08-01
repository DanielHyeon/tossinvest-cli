# Function Logic Map: `validateProtectionEvidence`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Untrusted descriptors for exactly two approved verifier tools. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 unknown/duplicate tool; B2 blank version or malformed build/evidence digest; B3 required tool absent; else metadata valid. | No mutation. Metadata validation alone is not authenticity or evidence verification. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `validSHA256`; no external reads. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Metadata validation alone is not authenticity or evidence verification.

## Safety conclusion

- Safe edit boundary: Require exact source basename, matrix binding digest, and later byte-for-byte SHA-256 recomputation; trusted signer remains blocker.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
