# Function Logic Map: `Journal.RecordFill`

Source: `internal/journal/fills.go`  
Function: `Journal.RecordFill`  
Signature: `Journal.RecordFill(params=2, results=2)`  
Source SHA-256: `8ee09a6b042e305d9e8d913eb86beb14f874d034f8ad8974ca488f8080699e9a`

## Inputs and invariants

- Inputs are the parameters in `Journal.RecordFill(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills.go:299 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/fills.go:303 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/fills.go:307 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/fills.go:319 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | switch | internal/journal/fills.go:326 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | case | internal/journal/fills.go:327 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | case | internal/journal/fills.go:328 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | case | internal/journal/fills.go:330 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/fills.go:336 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/fills.go:341 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/journal/fills.go:348 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/journal/fills.go:357 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | if | internal/journal/fills.go:361 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B14 | if | internal/journal/fills.go:369 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B15 | if | internal/journal/fills.go:376 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B16 | if | internal/journal/fills.go:377 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B17 | if | internal/journal/fills.go:393 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B18 | if | internal/journal/fills.go:397 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B19 | if | internal/journal/fills.go:398 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B20 | if | internal/journal/fills.go:413 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B21 | if | internal/journal/fills.go:414 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B22 | if | internal/journal/fills.go:452 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B23 | if | internal/journal/fills.go:460 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B24 | if | internal/journal/fills.go:464 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B25 | if | internal/journal/fills.go:476 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B26 | if | internal/journal/fills.go:477 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B27 | if | internal/journal/fills.go:482 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `strconv.ParseFloat`: returned errors and state follow the mapped branches.
- `orZero`: returned errors and state follow the mapped branches.
- `math.IsNaN`: returned errors and state follow the mapped branches.
- `math.IsInf`: returned errors and state follow the mapped branches.
- `j.nowString`: returned errors and state follow the mapped branches.
- `j.db.BeginTx`: returned errors and state follow the mapped branches.
- `tx.Rollback`: returned errors and state follow the mapped branches.
- `fillSnapshotScopeOf`: returned errors and state follow the mapped branches.
- `lookupFillSnapshotScoped`: returned errors and state follow the mapped branches.
- `errors.Is`: returned errors and state follow the mapped branches.
- `classifyFillRefusal`: returned errors and state follow the mapped branches.
- `markFillRefused`: returned errors and state follow the mapped branches.
- `alertsForOrder`: returned errors and state follow the mapped branches.
- `tx.Commit`: returned errors and state follow the mapped branches.
- `nearlyZero`: returned errors and state follow the mapped branches.
- `sameSnapshot`: returned errors and state follow the mapped branches.
- `upsertFillSnapshot`: returned errors and state follow the mapped branches.
- `recordExecutionCorrection`: returned errors and state follow the mapped branches.
- `firstNonEmpty`: returned errors and state follow the mapped branches.
- `tx.ExecContext`: returned errors and state follow the mapped branches.
- `decimalString`: returned errors and state follow the mapped branches.
- `resolveFillOrigin`: returned errors and state follow the mapped branches.
- `ownershipHandle.invalidate`: returned errors and state follow the mapped branches.
- `releaseReservationsForOrder`: returned errors and state follow the mapped branches.
- `fmt.Sprintf`: returned errors and state follow the mapped branches.
- `j.runApplyHooks`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 39 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
