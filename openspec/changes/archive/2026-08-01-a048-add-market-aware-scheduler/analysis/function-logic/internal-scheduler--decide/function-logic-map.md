# Function Logic Map: `Decide`

- Source: `internal/scheduler/decision.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| desired/activation | exact scheduler version, desired revision, approval fields and scope | `DesiredState` + sealed `Activation` | `DISABLED/NOT_ACTIVATED` |
| current calendar | non-nil, exact approved digest, same market, fresh and same exchange-local day | official adapter snapshot | `WAIT_MARKET` or `DISABLED` before budget spend |
| budget | endpoint-key coordinator | `BudgetCoordinator` | `BUDGET_DEFERRED` without entry poll |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | scheduler desired OFF | none | `DISABLED/SCHEDULER_OFF` | decision state matrix |
| B2 | sealed activation absent or desired binding mismatch | none | `DISABLED/NOT_ACTIVATED` | activation exact-state test |
| B3 | selected scope does not allow requested market | none | `DISABLED/MARKET_NOT_SELECTED` | decision state matrix |
| B4 | calendar missing | none | `WAIT_MARKET/CALENDAR_MISSING` | decision state matrix |
| B5 | calendar digest is empty or differs from both desired and activation | none | `DISABLED/NOT_ACTIVATED` | changed calendar version test |
| B6 | calendar market/freshness invalid | none | `WAIT_MARKET/CALENDAR_INVALID` | freshness tests |
| B7 | timezone/date mismatch | none | `WAIT_MARKET/CALENDAR_DAY_MISMATCH` | malformed/day boundary tests |
| B8 | holiday/no regular session | none | `WAIT_MARKET/HOLIDAY` | holiday test |
| B9-B10 | before open or at/after close | none | `WAIT_MARKET/MARKET_CLOSED` | decision state matrix |
| B11-B12 | coordinator absent or denies grant | coordinator may record no grant | `BUDGET_DEFERRED` | budget tests |
| success | all gates exact and grant allowed | consumes one entry grant | `ENTRY_ALLOWED` until close | decision state matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Activation.matches` | validate sealed desired/approval binding | pure; no retry | CodeGraph + AST |
| `CalendarSnapshot.ValidityAt` / `Market.Location` | freshness and exchange-local day | any uncertainty closes entry | CodeGraph + AST |
| `BudgetCoordinator.TryAcquire` | spend only discretionary entry budget | denial is returned, never retried here | CodeGraph + AST |

## State mutations and fallbacks

- Only the budget coordinator may mutate, and only after every authority/calendar gate passes.
- Calendar version mismatch is an authority mismatch, not a merely stale market wait.

## Safety conclusion

- Safe edit boundary: fail-closed entry decision only; exit/reconcile loops are outside this function.
- High-risk impact: yes, because `ENTRY_ALLOWED` is a future live-entry capability.
