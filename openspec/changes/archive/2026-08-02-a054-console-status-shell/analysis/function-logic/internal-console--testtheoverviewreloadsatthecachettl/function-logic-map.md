# Function Logic Map: `TestTheOverviewReloadsAtTheCacheTTL`

- Source: `internal/console/overview_test.go`
- AST evidence: `ast.json` (revision: current)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A test, updated for the moved screen. No assertion was weakened. The type-level assertion (overviewPage{}).Refresh() is gone because Refresh is now a field the handler sets. The render assertion that replaced it is stronger: it fails when a handler forgets to set it, which the method form could not detect.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 1074 | none; test code | unchanged behaviour | `TestTheOverviewReloadsAtTheCacheTTL` (this function is the test) |
| B2 | if at line 1085 | none; test code | unchanged behaviour | `TestTheOverviewReloadsAtTheCacheTTL` (this function is the test) |
| B3 | if at line 1088 | none; test code | unchanged behaviour | `TestTheOverviewReloadsAtTheCacheTTL` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 11 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None; test code.
- No new broker call, no new config key, no new audit record.

## Safety conclusion

- Safe edit boundary: display and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate, the ledger, reconciliation, authentication or fill handling. The console's set of state-changing acts does not grow: the two routes this change adds are GET reads with no form.
