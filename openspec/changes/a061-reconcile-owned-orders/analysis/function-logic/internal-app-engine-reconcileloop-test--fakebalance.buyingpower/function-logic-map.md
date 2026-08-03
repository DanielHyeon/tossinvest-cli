# Function Logic Map: `fakeBalance.BuyingPower`

Source: `internal/app/engine/reconcileloop_test.go`  
Function: `fakeBalance.BuyingPower`  
Signature: `fakeBalance.BuyingPower(params=2, results=2)`  
Source SHA-256: `f7244f04d716230ddc2536f8e219958c52b86a6b899cf6f4df45fa09962f961e`

## Inputs and invariants

- Inputs are the parameters in `fakeBalance.BuyingPower(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop_test.go:90 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- No outbound call; behavior is local and deterministic.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 0 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
