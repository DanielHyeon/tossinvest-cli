# Function Logic Map: `officialRawForPending`

- Source: `internal/verifylive/m0_recovery_test.go`; unchanged base fixture helper.
- AST evidence: `ast.json` (`revision: base`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| checkpoint and fixture ID | test fixture input | caller | returned in-memory raw conditional representation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unconditional fixture construction | none | returns raw fixture | `TestM0PendingRecoveryZeroOneMultipleHoldWithoutMutation` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | no calls | not applicable | AST |

## State mutations and fallbacks

- No mutation or external call; constructs test-only `official.RawConditionalOrder`.

## Safety conclusion

- Base fixture unchanged; map required because its containing test file changed.
