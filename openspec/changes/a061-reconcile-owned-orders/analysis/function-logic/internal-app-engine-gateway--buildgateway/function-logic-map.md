# Function Logic Map: `buildGateway`

Source: `internal/app/engine/gateway.go`  
Function: `buildGateway`  
Signature: `buildGateway(params=2, results=2)`  
Source SHA-256: `3dead101adcc3b89767975b14f72de7246909ac0ef3f909e3928ebed2637ee8b`

## Inputs and invariants

- Inputs are the parameters in `buildGateway(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/gateway.go:201 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/gateway.go:223 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/gateway.go:260 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `checkProjectionWired`: returned errors and state follow the mapped branches.
- `execgw.NewEntryGate`: returned errors and state follow the mapped branches.
- `tracker.Restore`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `entry.SetAuthorityRefresh`: returned errors and state follow the mapped branches.
- `tracker.Refresh`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `execgw.New`: returned errors and state follow the mapped branches.
- `in.official.BaseURL`: returned errors and state follow the mapped branches.
- `newNotifier`: returned errors and state follow the mapped branches.
- `newRetrier`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 13 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
