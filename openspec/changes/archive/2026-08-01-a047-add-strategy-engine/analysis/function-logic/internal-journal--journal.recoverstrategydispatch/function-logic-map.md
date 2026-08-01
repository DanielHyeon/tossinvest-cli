# Function Logic Map: `Journal.RecoverStrategyDispatch`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| pending receipt | account scoped PLANNED/IN_DOUBT exact receipt | `PendingStrategyPlans` | mismatch error |
| core attempts | bounded `LIMIT 2` rows for strategy intent id | execgw journal | exact terminal classification |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 569: `if receipt.AccountRef != accountRef \|\| (receipt.State != "PLANNED" && receipt.State != "IN_DOUBT") {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 573: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `for` at source line 579: `for rows.Next() {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 581: `if err := rows.Scan(&value.id, &value.state, &value.broker); err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 586: `if err := rows.Err(); err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 589: `if len(attempts) == 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 591: `if stateErr := j.RecordStrategyRefusal(ctx, receipt, "no_mutation_attempt"); stateErr != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `if` at source line 596: `if len(attempts) > 1 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B9 | exact AST `if` at source line 598: `if receipt.State == "PLANNED" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B10 | exact AST `if` at source line 599: `if stateErr := j.RecordStrategyInDoubt(ctx, receipt, "ambiguous_mutation_attempts"); stateErr != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B11 | exact AST `if` at source line 606: `if confirmed {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B12 | exact AST `if` at source line 609: `if attempts[0].state == string(StateNotDispatched) \|\| attempts[0].state == string(StateFailedConfirmed) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B13 | exact AST `if` at source line 611: `if stateErr := j.RecordStrategyRefusal(ctx, receipt, "mutation_attempt_refused"); stateErr != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B14 | exact AST `if` at source line 616: `if receipt.State == "PLANNED" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B15 | exact AST `if` at source line 617: `if stateErr := j.RecordStrategyInDoubt(ctx, receipt, "mutation_attempt_requires_recovery"); stateErr != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| bounded core query | detect 0/1/>1 without unbounded scan | DB errors propagate | AST + tests |
| terminal recorders | exact receipt CAS | never guess success | AST |

## State mutations and fallbacks

- Recovery performs no broker call and cannot create a second attempt.

## Safety conclusion

- Safe edit boundary: startup reconciliation of already durable plans.
- High-risk impact: yes, ambiguous exposure remains blocked.
