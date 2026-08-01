# Function Logic Map: `Journal.recordStrategyTerminal`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receipt | full exact plan binding and current revision/state | committed plan/startup enumeration | stale error |
| target/reason | REFUSED or IN_DOUBT, non-empty stable reason | dispatch/recovery classifier | invalid error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 626: `if j == nil \|\| j.db == nil \|\| receipt.AttemptID == "" \|\| receipt.AccountRef == "" \|\| receipt.DecisionIdentity == "" \|\| receipt.RiskIntentID == "" \|\| receipt.ClientOrderID == "" \|\| receipt.Quantity == "" \|\| receipt.Revision < 1 \|\| reason == "" \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 631: `if !allowedSource {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 635: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 640: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 643: `if storedState != receipt.State \|\| storedRevision != receipt.Revision {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 647: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 651: `if rows != 1 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `if` at source line 655: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyStrategyReceiptBindingTx` then state-only SQL CAS | bind account/decision/risk/client/quantity before updating only state/revision | rollback on mismatch | AST |
| refusal insert | preserve terminal reason history | same transaction | schema triggers/tests |

## State mutations and fallbacks

- PLANNED may become REFUSED/IN_DOUBT; IN_DOUBT may become REFUSED only with current receipt.

## Safety conclusion

- Safe edit boundary: exact strategy attempt terminal transitions.
- High-risk impact: yes, terminal truth must be durable and monotonic.
