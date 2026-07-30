# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- Qualified function: `TestSchemaTablesAndColumns`
- AST evidence: `ast.json` (`f168a7b83293e52443453b19c389ec3cb3740a2356b739381f172e4b55c4904b`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/schema_test.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/schema_test.go:112` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |
| B2 | `for` at `internal/journal/schema_test.go:116` — for rows.Next() { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |
| B3 | `if` at `internal/journal/schema_test.go:118` — if err := rows.Scan(&name); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |
| B4 | `if` at `internal/journal/schema_test.go:123` — if err := rows.Err(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |
| B5 | `if` at `internal/journal/schema_test.go:127` — if strings.Join(gotTables, ",") != strings.Join(wantTables, ",") { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |
| B6 | `range` at `internal/journal/schema_test.go:223` — for table, want := range wantColumns { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |
| B7 | `if` at `internal/journal/schema_test.go:226` — if strings.Join(got, ",") != strings.Join(want, ",") { | only mutations visible in this branch and its callees | existing return/error contract | `TestSchemaTablesAndColumns`; `TestSchemaV9AddsOnlyNullablePolicySnapshotColumns` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournal`, `context.Background`, `j.db.QueryContext`, `t.Fatal`, `rows.Next`, `rows.Scan`, `append`, `rows.Err`, `rows.Close`, `strings.Join`, `t.Fatalf`, `tableColumns` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 9 assignment(s) and 0 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
