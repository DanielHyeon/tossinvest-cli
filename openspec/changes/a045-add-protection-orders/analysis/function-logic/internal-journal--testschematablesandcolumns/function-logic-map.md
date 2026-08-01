# Function Logic Map: `TestSchemaTablesAndColumns`
- Source: `internal/journal/schema_test.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- The golden catalog lists every current journal table and exact columns for safety-critical ledgers.
## Branches and early returns
- B1-B7 cover catalog query failure, table scan iteration/error, exact table-list comparison, and exact per-table column comparison.
## Calls and live bindings
- Uses an isolated temporary journal and SQLite catalog/pragma queries.
## State mutations and fallbacks
- Test-only temporary database; any extra, missing, or renamed table/column fails the gate.
## Safety conclusion
- Pins the two additive v13 protection tables without weakening prior schema assertions.
