# Function Logic Map: `TestMigrationV12CommitAndUserVersionSurviveSIGKILL`
- Source: `internal/journal/migration_v12_crash_linux_test.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Legacy v12 crash fixture remains pinned to schema v12 while current schema advances to v13.
## Branches and early returns
- B1-B9 validate helper mode, process interruption, reopen, table survival, and exact user version.
## Calls and live bindings
- Uses temporary SQLite state and a subprocess crash boundary.
## State mutations and fallbacks
- Only temporary test database state is mutated; any partial migration fails the test.
## Safety conclusion
- Preserves historical crash-atomicity evidence after adding v13.
