# Function Logic Map: `Journal.RecordStrategyDispatched`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receipt | exact account/decision/risk/client/quantity/revision/state | committed strategy plan | reject stale/forged receipt |
| execution refs | non-empty core mutation and broker IDs | execgw Outcome | rollback on collision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 498: `if j == nil \|\| j.db == nil \|\| receipt.AttemptID == "" \|\| receipt.AccountRef != accountRef \|\| receipt.DecisionIdentity == "" \|\| receipt.RiskIntentID == "" \|\| receipt.ClientOrderID == "" \|\| receipt.Quantity == "" \|\| receipt.Revision < 1 \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 503: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 508: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 511: `if storedState != "DISPATCHED" \|\| storedRevision != receipt.Revision+1 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 512: `if storedState != receipt.State \|\| storedRevision != receipt.Revision {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 516: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 520: `if rows != 1 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `range` at source line 525: `for _, link := range [][2]string{{"MUTATION_ATTEMPT", mutationAttemptID}, {"BROKER_ORDER", brokerOrderID}} {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B9 | exact AST `if` at source line 527: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B10 | exact AST `if` at source line 531: `if err := tx.QueryRowContext(ctx, `SELECT attempt_id FROM strategy_execution_lineage WHERE account_ref=? AND kind=? AND external_ref=?`, accountRef, link[0], link[1]).Scan(&gotAttempt); err != nil \|\| gotAttempt != receipt.AttemptID {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyStrategyReceiptBindingTx` then state-only SQL CAS | exact immutable binding plus state/revision transition without naming immutable identity columns in UPDATE | one immediate transaction | AST |
| v14 unique indexes | one core attempt/order per strategy attempt | collision fails closed | schema tests |

## State mutations and fallbacks

- DISPATCHED never downgrades. IN_DOUBT can promote only with a current exact receipt.

## Safety conclusion

- Safe edit boundary: strategy terminal state and exact execution links only.
- High-risk impact: yes, records whether exposure may exist.
