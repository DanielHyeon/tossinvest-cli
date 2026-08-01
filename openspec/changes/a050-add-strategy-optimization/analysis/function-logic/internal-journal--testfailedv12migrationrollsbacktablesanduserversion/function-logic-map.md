# Function Logic Map: `TestFailedV12MigrationRollsBackTablesAndUserVersion`
- Source: `internal/journal/migration_v12_test.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Failure fixture intentionally targets migration 12, independent of current schema 13.
## Branches and early returns
- B1-B12 establish v11, inject a v12 failure, and assert version/table rollback and row preservation.
## Calls and live bindings
- Exercises SQLite migration transactions against a temporary database.
## State mutations and fallbacks
- Test database only; migration failure must be atomic.
## Safety conclusion
- Historical rollback test cannot accidentally migrate through the new protection schema.
