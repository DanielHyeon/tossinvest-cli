# Function Logic Map: `TestEveryRegisteredTierIsOfferedWithItsNumbers`

- Source: `internal/console/settings_limits_test.go`
- AST evidence: `ast.json` (revision: current)
- Change: a055-console-settings-cadence · category: test
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A test, re-pointed at the tab that owns the section it is about. No assertion was weakened; where a055's contract replaced an inherited one, the replacement is argued in the test's own comment.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at line 177 | none; test code | unchanged behaviour | `TestEveryRegisteredTierIsOfferedWithItsNumbers` (this function is the test) |
| B2 | if at line 178 | none; test code | unchanged behaviour | `TestEveryRegisteredTierIsOfferedWithItsNumbers` (this function is the test) |
| B3 | if at line 181 | none; test code | unchanged behaviour | `TestEveryRegisteredTierIsOfferedWithItsNumbers` (this function is the test) |
| B4 | if at line 184 | none; test code | unchanged behaviour | `TestEveryRegisteredTierIsOfferedWithItsNumbers` (this function is the test) |
| B5 | if at line 189 | none; test code | unchanged behaviour | `TestEveryRegisteredTierIsOfferedWithItsNumbers` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 12 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
