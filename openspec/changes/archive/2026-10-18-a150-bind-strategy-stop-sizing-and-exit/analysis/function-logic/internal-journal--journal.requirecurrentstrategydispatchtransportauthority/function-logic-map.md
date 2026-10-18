# Function Logic Map: `RequireCurrentStrategyDispatchTransportAuthority`

Base 28a800f5; source `internal/journal/strategy_dispatch_runtime.go`; authoritative branch and call extraction: `ast.json`; syntax cross-check: `ast-grep.txt`; scan: `risk-pattern-report.md`.

## Inputs and invariants

Inputs and durable identities are read from the function signature and callers. The approved design retains legacy v2/v3, authoritative risk and exit protection, atomic reservation, fail-closed v4, and no new network lock. Refer to `analysis/discovery-plan.md` and design decisions for exact binding.

## Branches and early returns

| Branch | Source line | Existing condition / branch | Required regression |
|---|---|---|---|
| B1 | 1660 | `if j == nil \|\| j.db == nil \|\| ctx == nil \|\| !validStrategyDispatchIdentity(request.LeaseID) \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B2 | 1666 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B3 | 1670 | `if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B4 | 1674 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B5 | 1677 | `if lease.State != StrategyDispatchLeaseSubmitting \|\| lease.Disposition != StrategyDispatchReservationReserved \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |

Return lines: 1663, 1667, 1671, 1675, 1680, 1682. Full expressions and effects are in source and `ast-grep.txt`; line excerpts do not replace source review.

## Calls and live bindings

AST calls: validStrategyDispatchIdentity, uint64, fmt.Errorf, j.db.BeginTx, tx.Rollback, requireCurrentStrategyDispatchOwner, loadStrategyDispatchLease, Before, UTC, j.clk.Now. Journal transaction failures return errors with rollback; projection and digest functions do not perform transport. Existing configuration binding and writer/replay scope are documented in discovery.

## State mutations and fallbacks

Inspect source and AST assignments for current mutations. No legacy fallback is permitted for marker 4. Existing protective exits and in-flight reservations remain authoritative; see approved design for transactional ordering and restrictions.

## Safety conclusion

High-risk impact: yes. Pre-edit inventory only; branch test outcomes are pending and cannot be inferred from these maps.

Source/function binding: `internal/journal/strategy_dispatch_runtime.go` / `Journal.RequireCurrentStrategyDispatchTransportAuthority`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
