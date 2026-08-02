# Function Logic Map: `Console.handleOrders`

- Source: `internal/console/orders_page.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

The page struct now embeds the shared `chrome`, so this handler builds it and the template reads the four status facts from it. What the handler reads, what it gates on and what it renders are unchanged. Passes the orders snapshot's own freshness to the strip and sets Refresh explicitly.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | no domain mutation; a GET render | unchanged behaviour | `TestEveryScreenRendersTheSameStatusStrip`, `TestEachScreenKeepsItsOwnReloadPeriod`, `TestEveryScreenIsCalledOneThing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 6 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; a get render.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
