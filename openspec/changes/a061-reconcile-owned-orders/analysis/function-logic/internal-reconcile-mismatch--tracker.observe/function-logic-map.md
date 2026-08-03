# Function Logic Map: `Tracker.Observe`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.Observe`  
Signature: `Tracker.Observe(params=2, results=2)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters in `Tracker.Observe(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:366 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/mismatch.go:373 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | else | internal/reconcile/mismatch.go:396 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | range | internal/reconcile/mismatch.go:380 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/reconcile/mismatch.go:381 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/mismatch.go:388 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | range | internal/reconcile/mismatch.go:398 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/reconcile/mismatch.go:399 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/reconcile/mismatch.go:405 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/reconcile/mismatch.go:418 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | range | internal/reconcile/mismatch.go:440 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | range | internal/reconcile/mismatch.go:443 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | if | internal/reconcile/mismatch.go:444 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B14 | range | internal/reconcile/mismatch.go:449 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B15 | range | internal/reconcile/mismatch.go:453 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B16 | range | internal/reconcile/mismatch.go:456 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B17 | if | internal/reconcile/mismatch.go:462 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B18 | else | internal/reconcile/mismatch.go:475 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B19 | range | internal/reconcile/mismatch.go:467 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B20 | if | internal/reconcile/mismatch.go:468 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `Now`: returned errors and state follow the mapped branches.
- `t.clock`: returned errors and state follow the mapped branches.
- `t.interval`: returned errors and state follow the mapped branches.
- `t.maxFailures`: returned errors and state follow the mapped branches.
- `t.mu.Lock`: returned errors and state follow the mapped branches.
- `diff.BlocksEntry`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `blocksFor`: returned errors and state follow the mapped branches.
- `block.Key`: returned errors and state follow the mapped branches.
- `fmt.Sprintf`: returned errors and state follow the mapped branches.
- `permanent.Key`: returned errors and state follow the mapped branches.
- `sortBlocks`: returned errors and state follow the mapped branches.
- `t.syncGate`: returned errors and state follow the mapped branches.
- `t.snapshotBlocks`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `t.persist`: returned errors and state follow the mapped branches.
- `delete`: returned errors and state follow the mapped branches.
- `hasPermanentQuantityAccountBlock`: returned errors and state follow the mapped branches.
- `now.Add`: returned errors and state follow the mapped branches.
- `t.mu.Unlock`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 39 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
