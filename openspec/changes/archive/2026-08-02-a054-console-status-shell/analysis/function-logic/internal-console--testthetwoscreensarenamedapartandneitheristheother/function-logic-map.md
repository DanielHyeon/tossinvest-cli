# Function Logic Map: `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther`

- Source: `internal/console/overview_test.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A test, updated for the moved screen. No assertion was weakened. The replacement. It keeps both original claims and adds one: a tab redirected off the old root must still be told a verification approval is waiting, with a link to the screen that can take it. That is the compensation for moving the screen.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 1138 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |
| B2 | if at line 1142 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |
| B3 | if at line 1145 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |
| B4 | if at line 1150 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |
| B5 | if at line 1153 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |
| B6 | if at line 1161 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |
| B7 | if at line 1164 | none; test code | unchanged behaviour | `TestTheTwoScreensAreNamedApartAndNeitherIsTheOther` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 22 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
