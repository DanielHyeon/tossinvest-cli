# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| migrated test journal | current production schema in isolated temp path | `openTestJournal` | open failure is fatal |
| expected tables/columns | exhaustive released schema contract | migration constants and typed readers | any missing/extra name fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | sqlite table query fails | none | fatal | same test |
| B2 | each table row scans | appends table name | scan failure fatal | same test |
| B3 | rows iterator fails | none | fatal | same test |
| B4 | exact table set differs | none | fatal | same test |
| B5 | rows close/check boundary | closes reader before column inspection | fatal on prior mismatch | same test |
| B6 | every expected table's columns are queried | read-only PRAGMA | loop | same test |
| B7 | exact column set differs/query fails | none | fatal | same test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournal` | opens current schema | fatal on error | AST |
| sqlite metadata queries | enumerate exact tables/columns | read only | AST |

## State mutations and fallbacks

- Read-only schema contract. a047 adds the three immutable v14 strategy-lineage tables to the exhaustive list.

## Safety conclusion

- Safe edit boundary: add exact v14 table names; keep all prior names and column assertions.
- High-risk impact: test-only verification of high-risk schema surface.
