# Function Logic Map: `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry`

- Source: `internal/console/optimization_review_block_test.go`
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
| B1 | if at line 184 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B2 | if at line 193 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B3 | if at line 201 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B4 | if at line 214 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B5 | range at line 218 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B6 | if at line 223 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B7 | range at line 227 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B8 | if at line 228 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B9 | if at line 233 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B10 | if at line 241 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B11 | if at line 250 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |
| B12 | if at line 254 | none; test code | unchanged behaviour | `TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 40 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
