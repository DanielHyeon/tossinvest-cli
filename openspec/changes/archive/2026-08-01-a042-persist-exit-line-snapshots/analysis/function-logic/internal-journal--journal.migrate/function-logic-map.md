# Function Logic Map: `Journal.migrate`

- Source: `internal/journal/journal.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current schema + ordered plan | no downgrade; backup before first live migration | journal migration contract | refusal/restore hint |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B9 | version read/newer/equal, backup, ordered steps, step failure, target check | backup metadata and per-step commit | error/success | migration v5-v10 suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| backup and applyMigration | crash-safe additive climb | one backup; step transaction | CodeGraph + AST |

## State mutations and fallbacks

- No planned body edit; v10 is appended to the immutable plan.

## Safety conclusion

- Safe edit boundary: unchanged migration runner.
- High-risk impact: yes.
