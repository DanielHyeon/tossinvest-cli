# Function Logic Map: `newProductionStrategyFirstLegAuthorityLoader`

Source `internal/app/engine/strategy_account_first_leg_authority.go`; base 28a800f5. AST `ast.json`; syntax `ast-grep.txt`; risk scan `risk-pattern-report.md`.

## Inputs and invariants

Exact authority and legacy compatibility from approved design. Preserve protection and reservations.

## Branches and early returns

| Branch | Line | Condition | Required regression |
|---|---|---|---|
| B1 | 277 | Branchless happy path | Preserve exact output. |

## Calls and live bindings

AST calls: detachedStrategyProposalPair. Source contains full expressions and error propagation; config binding is documented in discovery.

## State mutations and fallbacks

Journal transaction commits or rolls back atomically; do not infer missing v4 as legacy. Preserve existing protection.

## Safety conclusion

High-risk yes. Test mappings below are planned, not claims of executed coverage.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
