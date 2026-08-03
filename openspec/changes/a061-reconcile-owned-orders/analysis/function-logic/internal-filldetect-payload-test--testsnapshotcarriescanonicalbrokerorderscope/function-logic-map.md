# Function Logic Map: `TestSnapshotCarriesCanonicalBrokerOrderScope`

Source: `internal/filldetect/payload_test.go`  
Function: `TestSnapshotCarriesCanonicalBrokerOrderScope`  
Signature: `TestSnapshotCarriesCanonicalBrokerOrderScope(params=1, results=0)`  
Source SHA-256: `2a3179003b761f34a7ba63d94ba7f3c439689cc48bb00c04a23837a23f97fa9a`

## Inputs and invariants

- Inputs are the parameters in `TestSnapshotCarriesCanonicalBrokerOrderScope(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload_test.go:57 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `pollOne`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 1 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
