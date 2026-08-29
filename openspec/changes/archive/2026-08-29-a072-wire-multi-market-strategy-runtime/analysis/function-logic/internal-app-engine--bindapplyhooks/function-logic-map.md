# Function Logic Map: `bindApplyHooks`

- Source: `internal/app/engine/gateway.go`
- Qualified function: `bindApplyHooks`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| engine journal | non-nil, open journal that has not already bound hooks | engine construction order and `Journal.SetApplyHooks` | return error; engine construction refuses |
| hook set | Position, Campaign, Exit and Costs are installed as one literal | production engine wiring | partial or duplicate binding is rejected |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `SetApplyHooks` rejects nil/closed/duplicate binding | original hook set is preserved; no engine returned | wrapped binding error | `TestBindingWiresBothApplyHooks` and duplicate-binding engine tests |
| success | first binding | installs Position, Campaign, Exit and Costs together | nil | production static guard and ordinary/strategy fill suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.SetApplyHooks` | create the sole atomic fill-apply wiring point | called exactly once during construction; no retry/fallback | AST + engine wiring tests |
| `journal.ProjectPosition` | update authoritative position projection | runs inside fill transaction | journal hook rollback suites |
| `journal.ApplyPositionCampaignFill` | advance KR/US campaign lineage | runs inside the same fill transaction | paired strategy and ordinary fill tests |
| `journal.ApplyExitFill` | resolve exit state | runs after campaign in the same transaction | exit hook rollback suites |
| `costs.DefaultModel` | freeze the production cost model used by hook pricing | no runtime mutation | existing cost-model tests |

## State mutations and fallbacks

- This function mutates only the journal's in-memory one-time hook binding.
- Actual Position, Campaign and Exit writes occur later inside `Journal.RecordFill`'s transaction.
- There is no partial-hook or second-wiring fallback.

## Safety conclusion

- Safe edit boundary: keep all three domain appliers and Costs in the one `SetApplyHooks` literal.
- High-risk impact: yes — omitting Campaign or Exit would commit fills without complete lifecycle state.
