# Function Logic Map: `TestMigrationV14CommitAndUserVersionSurviveSIGKILL`

- Source: `internal/journal/migration_v14_crash_linux_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | test fixture and assertions | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:16`: `if os.Getenv(crashEnvMode) == crashModeMigrationV14AfterCommit {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B2 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:18`: `if version, err := j.SchemaVersion(context.Background()); err != nil \|\| version != 14 {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B3 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:28`: `if err := old.Close(); err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B4 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:34`: `if err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B5 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:39`: `if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil \|\| version != 14 {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B6 | AST `range` at `internal/journal/migration_v14_crash_linux_test.go:42`: `for _, name := range []string{"strategy_decision_lineage", "strategy_attempt_lineage", "strategy_execution_lineage", "strategy_attempt_refusals", "idx_strategy_execution_reverse"} ` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B7 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:44`: `if err := raw.QueryRow(\`SELECT count(*) FROM sqlite_master WHERE name=?\`, name).Scan(&count); err != nil \|\| count != 1 {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| B8 | AST `if` at `internal/journal/migration_v14_crash_linux_test.go:50`: `if after := countRows(t, reopened.db, v8AllTables); !sameCounts(before, after) {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.Getenv` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `openCrashChildJournalAtVersion` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `j.SchemaVersion` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `context.Background` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `t.Fatalf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `kill` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `filepath.Join` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |
| `t.TempDir` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/migration_v14_crash_linux_test.go` function `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` and its documented derived/test state.
- High-risk impact: no runtime authority.
