# Function Logic Map: `Journal.RecordFill`

Source: `internal/journal/fills.go`
Function: `Journal.RecordFill`
Signature: `Journal.RecordFill(params=2, results=2)`
Source SHA-256: `000918b94c8c3f776b611421c412e4604086fc4cbee2fd0e7c21fe0dd46454c0`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.RecordFill(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills.go:314 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/fills.go:318 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/fills.go:324 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/fills.go:328 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/fills.go:340 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | switch | internal/journal/fills.go:346 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | case | internal/journal/fills.go:347 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | case | internal/journal/fills.go:348 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | case | internal/journal/fills.go:350 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/fills.go:356 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/fills.go:364 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/fills.go:368 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/journal/fills.go:374 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | if | internal/journal/fills.go:380 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | if | internal/journal/fills.go:385 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | if | internal/journal/fills.go:392 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/journal/fills.go:401 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | if | internal/journal/fills.go:405 | Preserve the condition, error propagation, and fail-closed behavior. |
| B19 | if | internal/journal/fills.go:413 | Preserve the condition, error propagation, and fail-closed behavior. |
| B20 | if | internal/journal/fills.go:420 | Preserve the condition, error propagation, and fail-closed behavior. |
| B21 | if | internal/journal/fills.go:421 | Preserve the condition, error propagation, and fail-closed behavior. |
| B22 | if | internal/journal/fills.go:437 | Preserve the condition, error propagation, and fail-closed behavior. |
| B23 | if | internal/journal/fills.go:441 | Preserve the condition, error propagation, and fail-closed behavior. |
| B24 | if | internal/journal/fills.go:442 | Preserve the condition, error propagation, and fail-closed behavior. |
| B25 | if | internal/journal/fills.go:464 | Preserve the condition, error propagation, and fail-closed behavior. |
| B26 | if | internal/journal/fills.go:465 | Preserve the condition, error propagation, and fail-closed behavior. |
| B27 | if | internal/journal/fills.go:504 | Preserve the condition, error propagation, and fail-closed behavior. |
| B28 | if | internal/journal/fills.go:512 | Preserve the condition, error propagation, and fail-closed behavior. |
| B29 | if | internal/journal/fills.go:516 | Preserve the condition, error propagation, and fail-closed behavior. |
| B30 | if | internal/journal/fills.go:528 | Preserve the condition, error propagation, and fail-closed behavior. |
| B31 | if | internal/journal/fills.go:529 | Preserve the condition, error propagation, and fail-closed behavior. |
| B32 | if | internal/journal/fills.go:534 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `strings.TrimSpace`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `fillSnapshotScopeOf`: errors and state follow mapped branches.
- `scope.complete`: errors and state follow mapped branches.
- `scope.legacyUnscoped`: errors and state follow mapped branches.
- `strconv.ParseFloat`: errors and state follow mapped branches.
- `orZero`: errors and state follow mapped branches.
- `math.IsNaN`: errors and state follow mapped branches.
- `math.IsInf`: errors and state follow mapped branches.
- `j.nowString`: errors and state follow mapped branches.
- `j.db.BeginTx`: errors and state follow mapped branches.
- `tx.Rollback`: errors and state follow mapped branches.
- `lookupFillSnapshotScoped`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `confirmedFillOwners`: errors and state follow mapped branches.
- `journalTimeStrictlyAfter`: errors and state follow mapped branches.
- `classifyFillRefusal`: errors and state follow mapped branches.
- `markFillRefused`: errors and state follow mapped branches.
- `alertsForOrder`: errors and state follow mapped branches.
- `tx.Commit`: errors and state follow mapped branches.
- `nearlyZero`: errors and state follow mapped branches.
- `sameSnapshot`: errors and state follow mapped branches.
- `upsertFillSnapshot`: errors and state follow mapped branches.
- `recordExecutionCorrection`: errors and state follow mapped branches.
- `tx.ExecContext`: errors and state follow mapped branches.
- `decimalString`: errors and state follow mapped branches.
- `resolveFillOrigin`: errors and state follow mapped branches.
- `ownershipHandle.invalidate`: errors and state follow mapped branches.
- `releaseReservationsForOrder`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `j.runApplyHooks`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 45; return points: 22; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
