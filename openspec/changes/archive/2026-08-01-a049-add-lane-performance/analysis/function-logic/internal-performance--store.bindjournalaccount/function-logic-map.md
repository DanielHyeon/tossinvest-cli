# Function Logic Map: `Store.bindJournalAccount`

- Source: `internal/performance/lineage_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | validated journal-derived trades, observations, query/window, and derived-store state | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/performance/lineage_reader.go:76`: `if err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B2 | AST `if` at `internal/performance/lineage_reader.go:80`: `if _, err := tx.ExecContext(ctx, \`INSERT OR IGNORE INTO performance_scope(singleton,account_ref) VALUES(1,?)\`, strings.TrimSpace(accountRef)); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B3 | AST `if` at `internal/performance/lineage_reader.go:84`: `if err := tx.QueryRowContext(ctx, \`SELECT account_ref FROM performance_scope WHERE singleton=1\`).Scan(&bound); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B4 | AST `if` at `internal/performance/lineage_reader.go:87`: `if bound != strings.TrimSpace(accountRef) {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B5 | AST `if` at `internal/performance/lineage_reader.go:90`: `if err := tx.Commit(); err != nil {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.BeginTx` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `tx.Rollback` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `tx.ExecContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `strings.TrimSpace` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `Scan` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `tx.QueryRowContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `tx.Commit` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |

## State mutations and fallbacks

- local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/lineage_reader.go` function `Store.bindJournalAccount` and its documented derived/test state.
- High-risk impact: analytics only; no order, toggle, broker, or LIVE capability.
