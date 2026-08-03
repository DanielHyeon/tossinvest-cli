# Function Logic Map: `EntryGate.SetAuthorityRefresh`

Source: `internal/execgw/retry.go`  
Function: `EntryGate.SetAuthorityRefresh`  
Signature: `EntryGate.SetAuthorityRefresh(params=1, results=0)`  
Source SHA-256: `a549135ef2864ab05eb8168cfac899cad5052457189307c4ed3e3bee42e102d3`

## Inputs and invariants

- Inputs are the parameters in `EntryGate.SetAuthorityRefresh(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/execgw/retry.go:451 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `g.mu.Lock`: returned errors and state follow the mapped branches.
- `g.mu.Unlock`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 1 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
