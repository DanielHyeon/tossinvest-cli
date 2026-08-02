# Function Logic Map: `TestARestartedConsoleLabelsItselfAResume`

- Source: `internal/console/console_test.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A test, updated for the moved screen. No assertion was weakened. Opens the verification console at its new path. Same screen, same assertions.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at line 630 | none; test code | unchanged behaviour | `TestARestartedConsoleLabelsItselfAResume` (this function is the test) |
| B2 | if at line 631 | none; test code | unchanged behaviour | `TestARestartedConsoleLabelsItselfAResume` (this function is the test) |
| B3 | if at line 637 | none; test code | unchanged behaviour | `TestARestartedConsoleLabelsItselfAResume` (this function is the test) |
| B4 | if at line 640 | none; test code | unchanged behaviour | `TestARestartedConsoleLabelsItselfAResume` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 19 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
