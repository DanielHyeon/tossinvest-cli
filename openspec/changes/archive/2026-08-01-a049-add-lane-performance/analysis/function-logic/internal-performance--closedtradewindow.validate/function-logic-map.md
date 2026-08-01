# Function Logic Map: `ClosedTradeWindow.validate`

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
| B1 | AST `if` at `internal/performance/lineage_reader.go:31`: `if strings.TrimSpace(w.AccountRef) == "" {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B2 | AST `if` at `internal/performance/lineage_reader.go:34`: `if w.ClosedAfter.IsZero() \|\| w.ClosedAtOrBefore.IsZero() \|\| !w.ClosedAfter.Before(w.ClosedAtOrBefore) {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `errors.New` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `w.ClosedAfter.IsZero` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `w.ClosedAtOrBefore.IsZero` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| `w.ClosedAfter.Before` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |

## State mutations and fallbacks

- local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/lineage_reader.go` function `ClosedTradeWindow.validate` and its documented derived/test state.
- High-risk impact: analytics only; no order, toggle, broker, or LIVE capability.
