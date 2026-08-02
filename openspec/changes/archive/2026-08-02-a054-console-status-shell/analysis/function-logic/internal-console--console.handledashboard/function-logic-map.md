# Function Logic Map: `Console.handleDashboard`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A screen path or a landing path moved. Lost its path guard to redirectRoot and gained the shared chrome. The screen's own logic — snapshot, CSRF, restart seams, and Refresh only while a run is working and not awaiting — is byte-for-byte what it was.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 118 | no domain mutation; routing and rendering only | unchanged behaviour | `TestTheRootPathAnswersWithTheOverview`, `TestTheVerificationConsoleHasItsOwnPathAndKeepsItsControls`, `TestTheRestartHandoffSurvivesTheMove`, `TestARestartNoticeComesBackToTheScreenThatStartedIt`, `TestTheSessionCookieStaysScopedToTheWholeConsole` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 7 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; routing and rendering only.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
