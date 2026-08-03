# Function Logic Map: `exitContextForFill`

Source: `internal/journal/exit_state.go`  
Function: `exitContextForFill`  
Signature: `exitContextForFill(params=3, results=3)`  
Source SHA-256: `f3895fb41abc09f4de2aad1eceeeff1b39ab17ed658b2dc74e02bf7727b46f86`

## Inputs and invariants

- Inputs are the parameters in `exitContextForFill(params=3, results=3)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/exit_state.go:875 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/exit_state.go:878 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/exit_state.go:890 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/exit_state.go:895 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/exit_state.go:898 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `resolveFillOrigin`: returned errors and state follow the mapped branches.
- `tx.Query`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `instance.Close`: returned errors and state follow the mapped branches.
- `instance.Next`: returned errors and state follow the mapped branches.
- `instance.Err`: returned errors and state follow the mapped branches.
- `instance.Scan`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 4 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
