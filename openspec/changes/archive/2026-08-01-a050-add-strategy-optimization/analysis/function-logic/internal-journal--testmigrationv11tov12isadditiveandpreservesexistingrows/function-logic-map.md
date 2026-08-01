# Function Logic Map: `TestMigrationV11ToV12IsAdditiveAndPreservesExistingRows`
- Source: `internal/journal/migration_v12_test.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Historical v11-to-v12 behavior is explicitly bounded at v12 despite current v13.
## Branches and early returns
- B1-B6 set up v11, migrate one step, then verify version, old rows, and additive objects.
## Calls and live bindings
- Exercises migration helper and SQLite catalog queries.
## State mutations and fallbacks
- Temporary database only; assertion failures stop the gate.
## Safety conclusion
- Prevents v13 additions from invalidating the v12 compatibility contract.
