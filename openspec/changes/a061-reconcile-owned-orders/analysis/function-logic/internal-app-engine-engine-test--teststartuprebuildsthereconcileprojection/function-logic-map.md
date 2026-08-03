# Function Logic Map: `TestStartupRebuildsTheReconcileProjection`

Source: `internal/app/engine/engine_test.go`  
Function: `TestStartupRebuildsTheReconcileProjection`  
Signature: `TestStartupRebuildsTheReconcileProjection(params=1, results=0)`  
Source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

## Inputs and invariants

- Inputs are the parameters in `TestStartupRebuildsTheReconcileProjection(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:656 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/engine_test.go:661 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/engine_test.go:666 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/app/engine/engine_test.go:669 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/app/engine/engine_test.go:673 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/app/engine/engine_test.go:676 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `isolate`: returned errors and state follow the mapped branches.
- `writeEngineConfig`: returned errors and state follow the mapped branches.
- `writeCredentials`: returned errors and state follow the mapped branches.
- `engineStub`: returned errors and state follow the mapped branches.
- `seedAccountWideReconcile`: returned errors and state follow the mapped branches.
- `startEngine`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `eng.Entry.CheckEntry`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- `strings.Contains`: returned errors and state follow the mapped branches.
- `eng.Reconcile.Blocks`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `t.Error`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
