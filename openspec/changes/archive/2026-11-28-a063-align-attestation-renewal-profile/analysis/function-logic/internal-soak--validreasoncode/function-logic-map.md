# Function Logic Map: `validReasonCode`

- Source: `internal/soak/renewal_status.go` (93-99); SHA-256 `34fecab23a3018c1e9b076ea791939bbcd866657e14cad6fe25c624e745d5fc6` (post-edit). Pre-edit AST: `../../scope-absorb/pre-edit-ast/internal-soak--validreasoncode.json` (84-90, same 2 branches).
- AST evidence: `ast.json`. Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `code` | the closed `RenewalReasonCode` constants | constant block in `renewal_status.go` | false → caller rejects the status |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | switch on `code` | none | — | `TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason` |
| B2 | case: one of the closed codes — **D6 adds `ReasonLiveUncovered`** | none | true | `TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason` (refused + new code accepted) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | — | — | CodeGraph + AST |

## State mutations and fallbacks

- Pure. Fallthrough (no case) returns false: unknown codes stay invalid.

## Safety conclusion

- Safe edit boundary: one constant added to the accepted case list.
- High-risk impact: no.
