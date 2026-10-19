# Function Logic Map: `knownArmSuppression`

- Source: `internal/journal/exit_snapshot.go`
- Current and pre-edit AST: `ast.json`, `pre-edit-ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| reason | Three exact suppression constants | Journal caller and immutable snapshot validation | Unknown value rejected by validation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | working order, re-judge, or admitted-floor reconciliation | None | true; false for all others | Floor and snapshot persistence tests |

## Calls and live bindings

No calls. Snapshot validation uses this predicate before journal writes; it cannot turn suppression into an order.

## State mutations and fallbacks

Pure predicate; caller persists the reason separately from the evaluator snapshot.

## Safety conclusion

Only the explicit floor-reconciliation reason was added; existing reasons retain their meaning. Focused tests and reviews provide behavioral evidence separately.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
