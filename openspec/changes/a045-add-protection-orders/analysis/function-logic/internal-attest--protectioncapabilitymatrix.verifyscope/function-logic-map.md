# Function Logic Map: `ProtectionCapabilityMatrix.verifyScope`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Validated matrix and exact runtime account/profile/market/session/type/trigger/quantity/tool builds. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 no exact capability row; B2 tool set not exactly two; B3 version/build mismatch; else scope covered. | No mutation. Does not and must not change `ProtectionReady`. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | Account canonicalizer and typed equality checks. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Does not and must not change `ProtectionReady`.

## Safety conclusion

- Safe edit boundary: Compare strict canonical accounts and run only after evidence bytes and matrix binding verify.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
