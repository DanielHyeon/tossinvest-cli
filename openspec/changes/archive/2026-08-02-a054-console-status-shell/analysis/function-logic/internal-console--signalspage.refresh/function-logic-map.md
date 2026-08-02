# Function Logic Map: `signalsPage.Refresh`

- Source: `internal/console/signals.go`
- AST evidence: `ast.json` (revision: base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

Deleted. `Refresh` used to be a method on the page type, which reads as a property of the type; it is now a field on the embedded `chrome` that the handler sets. A screen cannot have both — the outer method silently shadows the promoted field — and one mechanism for "does this screen reload" is the point of the shared shell.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | none; the method had no body beyond a constant | unchanged behaviour | `TestEachScreenKeepsItsOwnReloadPeriod`, `TestTheReloadCellAndTheMetaTagAreOneFact` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 0 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; the method had no body beyond a constant.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
