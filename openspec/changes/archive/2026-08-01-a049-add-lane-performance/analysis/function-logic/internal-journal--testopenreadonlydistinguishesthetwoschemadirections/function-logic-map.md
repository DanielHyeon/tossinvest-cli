# Function Logic Map: `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`

- Source: `internal/journal/readonly_test.go`
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
| B1 | base-revision AST `if` at line 155: `if _, err := j.db.ExecContext(ctx, "PRAGMA user_version = 9999"); err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| B2 | base-revision AST `if` at line 160: `if !errors.Is(err, ErrSchemaTooNew) {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| B3 | base-revision AST `if` at line 177: `if err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| B4 | base-revision AST `if` at line 180: `if err := j.Close(); err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| B5 | base-revision AST `if` at line 185: `if !errors.Is(err, ErrSchemaTooOld) {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| B6 | base-revision AST `if` at line 190: `if !strings.Contains(err.Error(), "positions") {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `context.Background` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `t.Run` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `filepath.Join` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `t.TempDir` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `openTestJournalAt` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `j.db.ExecContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `t.Fatalf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |
| `OpenReadOnly` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/readonly_test.go` function `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` and its documented derived/test state.
- High-risk impact: no runtime authority.
