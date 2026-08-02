# Function Logic Map: `TestARefusedStartShowsTheEnginesOwnReason`

- Source: `internal/console/engineproc_test.go`
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
| B1 | if at line 150 | none; test code | unchanged behaviour | `TestARefusedStartShowsTheEnginesOwnReason` (this function is the test) |
| B2 | if at line 153 | none; test code | unchanged behaviour | `TestARefusedStartShowsTheEnginesOwnReason` (this function is the test) |
| B3 | if at line 158 | none; test code | unchanged behaviour | `TestARefusedStartShowsTheEnginesOwnReason` (this function is the test) |
| B4 | if at line 161 | none; test code | unchanged behaviour | `TestARefusedStartShowsTheEnginesOwnReason` (this function is the test) |
| B5 | if at line 165 | none; test code | unchanged behaviour | `TestARefusedStartShowsTheEnginesOwnReason` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 14 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
