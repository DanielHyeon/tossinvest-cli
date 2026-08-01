# Function Logic Map: `ReadOnly.checkSchema`

- Source: `internal/journal/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | context, journal/transaction state, persisted lineage and schema version | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/journal/readonly.go:203`: `if err := r.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&r.version); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B2 | AST `if` at `internal/journal/readonly.go:206`: `if r.version > SchemaVersion {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B3 | AST `range` at `internal/journal/readonly.go:212`: `for _, table := range readOnlyTables {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B4 | AST `switch` at `internal/journal/readonly.go:216`: `switch {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B5 | AST `case` at `internal/journal/readonly.go:217`: `case errors.Is(err, sql.ErrNoRows):` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B6 | AST `case` at `internal/journal/readonly.go:219`: `case err != nil:` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B7 | AST `if` at `internal/journal/readonly.go:223`: `if len(missing) > 0 {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B8 | AST `range` at `internal/journal/readonly.go:228`: `for _, required := range readOnlyColumns {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B9 | AST `switch` at `internal/journal/readonly.go:233`: `switch {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B10 | AST `case` at `internal/journal/readonly.go:234`: `case errors.Is(err, sql.ErrNoRows):` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B11 | AST `case` at `internal/journal/readonly.go:236`: `case err != nil:` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| B12 | AST `if` at `internal/journal/readonly.go:240`: `if len(missing) > 0 {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Scan` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| `r.db.QueryRowContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| `errors.Is` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| `append` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| `len` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |
| `strings.Join` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`, `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`, `TestTheReadOnlyHandleHasNoWriteMethods` |

## State mutations and fallbacks

- SQLite transaction or read state only; errors and missing evidence fail closed.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/readonly.go` function `ReadOnly.checkSchema` and its documented derived/test state.
- High-risk impact: journal correctness is high-risk, but this function has no broker/order capability.
