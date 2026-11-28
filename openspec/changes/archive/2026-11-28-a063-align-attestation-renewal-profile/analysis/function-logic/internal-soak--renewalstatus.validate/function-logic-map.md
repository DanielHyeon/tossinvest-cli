# Function Logic Map: `RenewalStatus.Validate`

- Source: `internal/soak/renewal_status.go` (56-81); SHA-256 `34fecab23a3018c1e9b076ea791939bbcd866657e14cad6fe25c624e745d5fc6` (post-edit). Pre-edit AST: `../../scope-absorb/pre-edit-ast/internal-soak--renewalstatus.validate.json` (HEAD `baeddfe31`, 53-77, same 7 branches).
- AST evidence: `ast.json`. Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` (status), `now` | format 1, UTC attempt not after `now`, closed outcome and codes, ≤16 unique codes | `SaveRenewalStatus` (writer) and `LoadRenewalStatus` (reader) both call it | fixed `invalid renewal status` error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | bad version / zero, future or non-UTC attempt | none | error | `TestRenewalStatusRejectsEmptyFailureReasonsAndNonUTCTimestamps` |
| B2 | outcome not issued/refused/failed | none | error | `TestRenewalStatusRoundTripAndRejectsUnsafeInputs` |
| B3 | more than 16 codes | none | error | `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` (bounded schema family) |
| B4 | range over codes | builds `seen` | — | `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` |
| B5 | duplicate or unknown code | none | error | `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` |
| B6 | issued AND (reasons not allowed by `issuedReasonsAllowed` — anything but none or exactly `[live_endpoint_uncovered]` — OR zero/non-UTC expiry OR expiry not after attempt). **D6 edit: the first operand was `!s.ReasonCodesEmpty()`.** | none | error | `TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason`, `TestRenewalStatusIssuedNeedsConsistentExpiry` |
| B7 | non-issued without codes | none | error | `TestRenewalStatusRejectsEmptyFailureReasonsAndNonUTCTimestamps` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `isUTCOffset`, `validReasonCode`, `RenewalStatus.issuedReasonsAllowed` (new, D6) | pure predicates | none | CodeGraph + AST |

## State mutations and fallbacks

- Pure; no I/O. A rejected status is never written (writer) and reads as unknown in the console (reader).

## Safety conclusion

- Safe edit boundary: only B6's reason-code operand widened from "none" to "none or the single D6 code"; every other predicate unchanged.
- High-risk impact: no — diagnostics only; the engine and startup interlock never read this file.
