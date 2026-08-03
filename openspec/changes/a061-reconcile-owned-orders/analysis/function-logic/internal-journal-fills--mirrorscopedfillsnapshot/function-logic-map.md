# Function Logic Map: `mirrorScopedFillSnapshot`

Source: `internal/journal/fills.go`  
Function: `mirrorScopedFillSnapshot`  
Signature: `mirrorScopedFillSnapshot(params=8, results=1)`  
Source SHA-256: `8ee09a6b042e305d9e8d913eb86beb14f874d034f8ad8974ca488f8080699e9a`

## Inputs and invariants

- Inputs are the parameters in `mirrorScopedFillSnapshot(params=8, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/fills.go:608 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `tx.ExecContext`: returned errors and state follow the mapped branches.
- `boolToInt`: returned errors and state follow the mapped branches.
- `orZero`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 1 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
