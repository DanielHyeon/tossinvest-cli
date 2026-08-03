# Function Logic Map: `newEngineCmd`

Source: `cmd/tossctl/engine.go`  
Function: `newEngineCmd`  
Signature: `newEngineCmd(params=1, results=1)`  
Source SHA-256: `45414562be8a352d2183fb2dfc0985154e0eea5ce781e167eb6800841c495451`

## Inputs and invariants

- Inputs are the parameters in `newEngineCmd(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | cmd/tossctl/engine.go:112 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `cmd.AddCommand`: returned errors and state follow the mapped branches.
- `newEngineRunCmd`: returned errors and state follow the mapped branches.
- `newEngineReconcileResolveCmd`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 1 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
