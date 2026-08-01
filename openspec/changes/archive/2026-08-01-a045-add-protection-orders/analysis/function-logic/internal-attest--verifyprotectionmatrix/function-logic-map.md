# Function Logic Map: `verifyProtectionMatrix`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authenticated parsed matrix and runtime scope/evidence | non-empty matrix plus exact scope and evidence | signed payload + runtime caller | typed invalid/scope/evidence error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | parsed matrix is empty | none | `ErrProtectionInvalid` | empty parsed test |
| B2 | no exact scope row | none | `ErrProtectionScope` | replay/mismatch table |
| B3 | exact evidence set/bytes mismatch | potentially large digest computation only | `ErrProtectionEvidence` | missing/extra/tamper table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyScope`, `verifyProtectionEvidenceBytes` | finish all caller-controlled or unbounded work before the final trust-root read | pure, no retry/fallback | CodeGraph + AST |

## State mutations and fallbacks

- No mutation; returns only the matched row to the final verifier phase and creates no authority.

## Safety conclusion

- Safe edit boundary: prevalidates exact runtime claims/evidence before current revocation state is sampled.
- High-risk impact: yes; authority minimization boundary.
