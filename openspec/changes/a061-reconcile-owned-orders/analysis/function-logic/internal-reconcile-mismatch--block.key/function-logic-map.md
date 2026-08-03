# Function Logic Map: `Block.Key`

Source: `internal/reconcile/mismatch.go`  
Function: `Block.Key`  
Signature: `Block.Key(params=0, results=1)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters in `Block.Key(params=0, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/reconcile/mismatch.go:223 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | case | internal/reconcile/mismatch.go:224 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | case | internal/reconcile/mismatch.go:226 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | case | internal/reconcile/mismatch.go:228 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | case | internal/reconcile/mismatch.go:230 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `string`: returned errors and state follow the mapped branches.
- `strings.ToLower`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 0 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
