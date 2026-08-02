# Function Logic Map: `TestTrustedNetworkConsoleNeedsNoApplicationSession`

- Source: `internal/console/remote_test.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A test, updated for the moved screen. No assertion was weakened. Fetches the overview instead of the root. The root is now always a 303, and a request that always redirects cannot distinguish an authorised session (200) from an unauthorised one (303 to /login) — which is the whole assertion. Nothing else changed.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 181 | none; test code | unchanged behaviour | `TestTrustedNetworkConsoleNeedsNoApplicationSession` (this function is the test) |
| B2 | if at line 184 | none; test code | unchanged behaviour | `TestTrustedNetworkConsoleNeedsNoApplicationSession` (this function is the test) |
| B3 | range at line 187 | none; test code | unchanged behaviour | `TestTrustedNetworkConsoleNeedsNoApplicationSession` (this function is the test) |
| B4 | if at line 188 | none; test code | unchanged behaviour | `TestTrustedNetworkConsoleNeedsNoApplicationSession` (this function is the test) |
| B5 | if at line 192 | none; test code | unchanged behaviour | `TestTrustedNetworkConsoleNeedsNoApplicationSession` (this function is the test) |
| B6 | if at line 198 | none; test code | unchanged behaviour | `TestTrustedNetworkConsoleNeedsNoApplicationSession` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 18 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
