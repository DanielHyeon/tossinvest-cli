# Function Logic Map: `Journal.applyMigration`

- Source: `internal/journal/journal.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one migration step | SQL, metadata, user_version commit together | journal schema contract | rollback on any error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | begin, DDL, metadata stamps, user_version, commit | one schema transaction | wrapped error/success | v10 fault matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite transaction | atomic schema step | rollback deferred | CodeGraph + AST |

## State mutations and fallbacks

- No planned body edit; v10 fault tests exercise the existing transaction.

## Safety conclusion

- Safe edit boundary: unchanged.
- High-risk impact: yes.
