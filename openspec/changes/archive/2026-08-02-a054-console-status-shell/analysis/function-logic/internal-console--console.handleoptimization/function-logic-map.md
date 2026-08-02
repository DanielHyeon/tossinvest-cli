# Function Logic Map: `Console.handleOptimization`

- Source: `internal/console/optimization.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

The page struct now embeds the shared `chrome`, so this handler builds it and the template reads the four status facts from it. What the handler reads, what it gates on and what it renders are unchanged.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 66 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B2 | if at line 73 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B3 | if at line 80 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B4 | if at line 85 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B5 | if at line 87 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B6 | else at line 89 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B7 | if at line 93 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B8 | if at line 95 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B9 | else at line 97 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B10 | range at line 106 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B11 | if at line 108 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B12 | if at line 115 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B13 | if at line 118 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B14 | else at line 120 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 28 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; a get render.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
