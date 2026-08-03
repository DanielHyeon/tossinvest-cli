# Function Logic Map: `engineFillDetector`

Source: `cmd/tossctl/engine.go`  
Function: `engineFillDetector`  
Signature: `engineFillDetector(params=3, results=2)`  
Source SHA-256: `45414562be8a352d2183fb2dfc0985154e0eea5ce781e167eb6800841c495451`

## Inputs and invariants

- Inputs are the parameters in `engineFillDetector(params=3, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | cmd/tossctl/engine.go:401 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | cmd/tossctl/engine.go:402 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `hints.Validate`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `ectx.AccountSweep`: returned errors and state follow the mapped branches.
- `ectx.SnapshotCurrencies`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 3 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
