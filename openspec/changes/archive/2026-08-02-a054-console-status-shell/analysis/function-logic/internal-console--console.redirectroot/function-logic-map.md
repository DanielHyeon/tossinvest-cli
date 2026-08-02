# Function Logic Map: `Console.redirectRoot`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A screen path or a landing path moved. New function split out of handleDashboard. It keeps handleDashboard's 404 guard for every unserved path (this is the mux catch-all) and answers the root itself with a 303 to the overview instead of rendering the verification console.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 100 | no domain mutation; routing and rendering only | unchanged behaviour | `TestTheRootPathAnswersWithTheOverview`, `TestTheVerificationConsoleHasItsOwnPathAndKeepsItsControls`, `TestTheRestartHandoffSurvivesTheMove`, `TestARestartNoticeComesBackToTheScreenThatStartedIt`, `TestTheSessionCookieStaysScopedToTheWholeConsole` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 2 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; routing and rendering only.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
