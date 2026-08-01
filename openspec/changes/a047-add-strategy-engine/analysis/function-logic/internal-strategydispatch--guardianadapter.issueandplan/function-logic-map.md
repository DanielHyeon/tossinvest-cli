# Function Logic Map: `GuardianAdapter.IssueAndPlan`

- Source: `internal/strategydispatch/adapters.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| concrete adapters | Guardian, journal, account reader, collector non-nil | dormant approved wiring | reject unconfigured |
| manifest binding | account/policy/limits/settings exact | Guardian + activation | reject mismatch |
| atomic receipt | attempt/account/decision/risk/client/quantity/revision/state exact | committed journal result | reject mismatch, never post-read |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 30: `if a == nil \|\| a.Guardian == nil \|\| a.Journal == nil \|\| a.Account == nil \|\| a.Collect == nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 33: `if request.Binding.AccountRef != a.Guardian.AccountRef() \|\| request.Binding.GuardianVersion != a.Guardian.PolicyVersion() \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 38: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 46: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 50: `if receipt.AttemptID != request.AttemptID \|\| receipt.AccountRef != request.Binding.AccountRef \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RiskGuardian.IssueStrategyEntry` | sole strategy authority issuer | exact sizing + atomic commit | AST |
| committed receipt | avoid fallible post-commit lookup | canonical client id/quantity included | journal tests |

## State mutations and fallbacks

- Does not call the gateway. A returned plan is already durable and recoverable.

## Safety conclusion

- Safe edit boundary: concrete bridge only; no fake Authorize/Plan split.
- High-risk impact: yes, transfers committed authority to gateway plan.
