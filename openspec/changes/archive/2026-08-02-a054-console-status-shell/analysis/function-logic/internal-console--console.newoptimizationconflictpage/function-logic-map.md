# Function Logic Map: `Console.newOptimizationConflictPage`

- Source: `internal/console/optimization_view.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A page constructor now reaches the console so it can build the shared chrome. The free newOptimizationConflictPage, now a method so it can build the shared chrome. The conflict-row construction and the re-preview eligibility rule are copied unchanged.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at line 177 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B2 | if at line 179 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B3 | if at line 187 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B4 | if at line 193 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 9 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; a get render.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
