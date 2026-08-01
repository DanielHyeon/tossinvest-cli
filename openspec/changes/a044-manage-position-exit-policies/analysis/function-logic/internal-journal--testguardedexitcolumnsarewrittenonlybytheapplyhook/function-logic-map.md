# Function Logic Map: `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`

- Source: `internal/journal/apply_hook_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| production journal sources | all non-test Go files | package directory | read failure is fatal |\n| guarded columns | exactly four fill-time fields | `guardedColumns` | test fails on drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B10 | walk/read/allowlist/update scans | none | test failure on extra reader/writer | this test |\n| B11-B13 | positive controls | none | test failure if scan is vacuous | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `productionSources`/`updateStatements` | lexical defense-in-depth | deterministic local scan | AST |

## State mutations and fallbacks

- Test-only. `position_policy.go` may read pending status, while any UPDATE outside `apply_hook.go` remains rejected.

## Safety conclusion

- Safe edit boundary: separate reader allowlisting from the sole-writer invariant so list projection cannot weaken atomic fill ownership.
- High-risk impact: yes
