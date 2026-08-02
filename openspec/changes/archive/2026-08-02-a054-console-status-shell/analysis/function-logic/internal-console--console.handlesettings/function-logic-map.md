# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

The page struct now embeds the shared `chrome`, so this handler builds it and the template reads the four status facts from it. What the handler reads, what it gates on and what it renders are unchanged. Reads at render time. Every seam check, load and gate in the body is unchanged.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 147 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B2 | if at line 150 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B3 | if at line 155 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B4 | if at line 158 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B5 | if at line 163 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B6 | if at line 166 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B7 | if at line 175 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B8 | if at line 178 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |
| B9 | if at line 184 | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 16 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; a get render.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
