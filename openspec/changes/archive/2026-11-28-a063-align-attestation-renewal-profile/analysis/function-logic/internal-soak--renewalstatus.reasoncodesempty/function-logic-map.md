# Function Logic Map: `RenewalStatus.ReasonCodesEmpty`

- Source: `internal/soak/renewal_status.go` (87-87); SHA-256 `34fecab23a3018c1e9b076ea791939bbcd866657e14cad6fe25c624e745d5fc6` (post-edit current).
- AST evidence: `ast.json` (branchless). Risk scan: `risk-pattern-report.md`.
- Why this bundle exists: the body is **unchanged** by a063 D6. Commit `8b5ca6edd` inserted the new helper `RenewalStatus.issuedReasonsAllowed` directly after `ReasonCodesEmpty` (and the const comment shifted lines), so the `--unified=0` hunk touches this function's span and the a125 receipt attributes it. Recorded honestly rather than rewriting history.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.ReasonCodes` | any | caller `RenewalStatus.Validate` | pure boolean |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy path | none | boolean | `TestRenewalStatusRejectsEmptyFailureReasonsAndNonUTCTimestamps` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| standard library only | — | none | CodeGraph + AST |

## State mutations and fallbacks

- Pure; no I/O.

## Safety conclusion

- Safe edit boundary: no edit to this body.
- High-risk impact: no.
