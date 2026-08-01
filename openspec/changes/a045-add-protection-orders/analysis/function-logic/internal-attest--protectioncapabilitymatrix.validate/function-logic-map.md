# Function Logic Map: `ProtectionCapabilityMatrix.validate`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Decoded but untrusted versioned matrix and verification time. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | version mismatch, invalid/non-UTC time window, invalid/unsorted evidence, no rows, invalid/unsorted/duplicate rows, marshal or digest mismatch. | No mutation; canonical digest binds the semantic matrix. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `validateProtectionEvidence`, `ConditionalCapability.validate`; pure except time comparison. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation or fallback; only one sorted UTC canonical semantic form is accepted.

## Safety conclusion

- Safe edit boundary: Validate an explicit digest over canonical capability rows and require every evidence descriptor to bind that digest.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
