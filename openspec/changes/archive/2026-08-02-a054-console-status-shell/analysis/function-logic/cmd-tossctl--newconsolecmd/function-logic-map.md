# Function Logic Map: `newConsoleCmd`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A screen path or a landing path moved. Help text only: the verification console is described at its new path and the root is described as a redirect. No flag, seam or wiring changed.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | no domain mutation; routing and rendering only | unchanged behaviour | `TestTheRootPathAnswersWithTheOverview`, `TestTheVerificationConsoleHasItsOwnPathAndKeepsItsControls`, `TestTheRestartHandoffSurvivesTheMove`, `TestARestartNoticeComesBackToTheScreenThatStartedIt`, `TestTheSessionCookieStaysScopedToTheWholeConsole` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 18 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; routing and rendering only.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
