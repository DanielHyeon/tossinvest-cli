# Function Logic Map: `enterReconcileInTx`

Source: `internal/journal/position_projection.go`  
Function: `enterReconcileInTx`  
Signature: `enterReconcileInTx(params=5, results=1)`  
Source SHA-256: `ae74d3ba1b66a05360e7b5851248fd6814577fa0b34068a89f52c58c10644c7b`

## Inputs and invariants

- Inputs are the parameters in `enterReconcileInTx(params=5, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/position_projection.go:369 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `reconcileCauseFor`: returned errors and state follow the mapped branches.
- `fmt.Sprintf`: returned errors and state follow the mapped branches.
- `enterReconcileScopeInTx`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 2 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
