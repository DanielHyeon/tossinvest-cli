# Function Logic Map: `ReadOnly.checkSchema`

- Source: `internal/journal/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.db` | existing SQLite handle opened `mode=ro`, `query_only` | `OpenReadOnly` | close handle and return typed/schema error |
| `PRAGMA user_version` | `<= SchemaVersion`; released prerequisites present | journal migration lineage | `ErrSchemaTooNew` or `ErrSchemaTooOld` |
| v21 evidence columns | both nullable columns present only when version >=21 | a064 migration | `ErrSchemaTooOld`, never migrate/write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | user_version read error or newer schema | only receiver version scan | wrapped error / `ErrSchemaTooNew` | existing schema-direction tests |
| B3-B7 | required base table missing/query fails/aggregate nonempty | append local missing names only | query error / `ErrSchemaTooOld` | existing readonly table tests |
| B8-B12 | required released column missing/query fails/aggregate nonempty | append local missing names only | query error / `ErrSchemaTooOld` | existing v8/v9/v15 tests |
| B13-B22 | v20 campaign table/column checks | append local missing names only | query error / `ErrSchemaTooOld` | a065 damaged-v20 tests |
| B23-B28 (new) | version >=21 and consumed snapshot ID/digest column missing/query fails | append local missing names only | query error / `ErrSchemaTooOld` | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` |
| success | all prerequisites for observed version exist | no write | nil | readonly lineage read test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `QueryRowContext(...).Scan` | inspect version, tables and columns without migration | caller context bounds the read; any query error is returned | CodeGraph caller `OpenReadOnly` + AST |
| `errors.Is(sql.ErrNoRows)` | distinguish absent schema artifact from driver failure | absence accumulates into typed old-schema diagnostic | existing readonly tests |

## State mutations and fallbacks

- No SQL mutation exists; the handle remains structurally read-only.
- Released diagnostic ordering remains base columns, then v20 campaign, then v21 snapshot lineage.
- v20 and older readers do not demand future columns; v21 readers fail closed if either nullable column is damaged.

## Safety conclusion

- Safe edit boundary: append one version-gated column inspection after existing v20 validation.
- High-risk impact: yes — journal schema validation, mitigated by SELECT-only access and typed fail-closed tests.
