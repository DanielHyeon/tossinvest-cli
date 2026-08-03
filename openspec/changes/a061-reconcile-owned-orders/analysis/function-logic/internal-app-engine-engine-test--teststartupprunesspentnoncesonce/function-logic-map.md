# Function Logic Map: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`  
Function: `TestStartupPrunesSpentNoncesOnce`  
Signature: `TestStartupPrunesSpentNoncesOnce(params=1, results=0)`  
Source SHA-256: `dec52090cdbaa17f2a868d8e204c1755d88ec127a3f5158efc858405f41b83b7`

## Inputs and invariants

- Inputs are the parameters in `TestStartupPrunesSpentNoncesOnce(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:451 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/engine_test.go:456 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/engine_test.go:460 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `isolate`: returned errors and state follow the mapped branches.
- `writeEngineConfig`: returned errors and state follow the mapped branches.
- `writeCredentials`: returned errors and state follow the mapped branches.
- `engineStub`: returned errors and state follow the mapped branches.
- `time.Now`: returned errors and state follow the mapped branches.
- `seedSpentNonce`: returned errors and state follow the mapped branches.
- `now.Add`: returned errors and state follow the mapped branches.
- `startEngine`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `eng.Journal.NonceSpent`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 9 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
