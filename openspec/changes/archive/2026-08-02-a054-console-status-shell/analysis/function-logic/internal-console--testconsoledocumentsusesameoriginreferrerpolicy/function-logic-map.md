# Function Logic Map: `TestConsoleDocumentsUseSameOriginReferrerPolicy`

- Source: `internal/console/referrer_policy_test.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A test, updated for the moved screen. No assertion was weakened. Page struct construction only.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at line 28 | none; test code | unchanged behaviour | `TestConsoleDocumentsUseSameOriginReferrerPolicy` (this function is the test) |
| B2 | if at line 33 | none; test code | unchanged behaviour | `TestConsoleDocumentsUseSameOriginReferrerPolicy` (this function is the test) |
| B3 | if at line 36 | none; test code | unchanged behaviour | `TestConsoleDocumentsUseSameOriginReferrerPolicy` (this function is the test) |
| B4 | if at line 39 | none; test code | unchanged behaviour | `TestConsoleDocumentsUseSameOriginReferrerPolicy` (this function is the test) |
| B5 | else at line 42 | none; test code | unchanged behaviour | `TestConsoleDocumentsUseSameOriginReferrerPolicy` (this function is the test) |
| B6 | if at line 42 | none; test code | unchanged behaviour | `TestConsoleDocumentsUseSameOriginReferrerPolicy` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 13 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
