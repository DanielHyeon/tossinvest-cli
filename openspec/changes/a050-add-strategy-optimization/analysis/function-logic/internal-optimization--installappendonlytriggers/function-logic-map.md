# Function Logic Map: `installAppendOnlyTriggers`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| migration transaction | active schema transaction | `Store.migrate` | first trigger install error aborts transaction |
| trigger manifest | fixed update/delete refusal triggers for all immutable tables | `appendOnlyTriggers` | no partial committed trigger set |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate every trigger definition | transaction-local DDL | continue | schema append-only test |
| B2 | trigger creation fails | transaction-local only | wrapped trigger-specific error | migration rollback test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| transaction `ExecContext` | installs idempotent manifest triggers | one statement each; migration transaction rolls back on error | schema/rollback tests |

## State mutations and fallbacks

- DDL is confined to the caller transaction. `IF NOT EXISTS` is followed by exact-definition verification, so it cannot bless a drifted same-name trigger.

## Safety conclusion

- Safe edit boundary: append-only schema protections.
- High-risk impact: yes; audit/snapshot/candidate/application history must remain immutable.
