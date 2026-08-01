# Function Logic Map: `Store.CollectClosedStrategyTrades`

- Source: `internal/performance/lineage_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| reader | exact persisted-ID lineage reader, non-nil | later journal adapter handoff | nil/read error fails before writes |
| observations | caller-owned map keyed by exact position ID | caller | no fetch/poll fallback |
| replay | reader may return already collected closes | immutable store | exact bytes skip; divergent bytes fail |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at line 50: `if reader == nil {` | derived-store transaction(s) only after account/window/source validation | explicit success/error/continue path; no invented fallback | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B2 | AST `if` at line 53: `if err := window.validate(); err != nil {` | derived-store transaction(s) only after account/window/source validation | explicit success/error/continue path; no invented fallback | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B3 | AST `if` at line 57: `if err != nil {` | derived-store transaction(s) only after account/window/source validation | explicit success/error/continue path; no invented fallback | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B4 | AST `if` at line 60: `if err := s.bindJournalAccount(ctx, window.AccountRef); err != nil {` | derived-store transaction(s) only after account/window/source validation | explicit success/error/continue path; no invented fallback | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B5 | AST `range` at line 64: `for _, trade := range trades {` | derived-store transaction(s) only after account/window/source validation | explicit success/error/continue path; no invented fallback | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |
| B6 | AST `if` at line 66: `if err != nil {` | derived-store transaction(s) only after account/window/source validation | explicit success/error/continue path; no invented fallback | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ClosedStrategyTrades` | one exact journal-derived batch read | no approximate API | current HEAD + interface tests |
| `Store.Collect` | compare-and-appends one trade atomically | exact replay accepted, divergence refused | store tests |

## State mutations and fallbacks

- The seam has no symbol/time nearest-neighbour method and receives observation values, not a polling capability.
- Journal schema adapter remains intentionally unwired until a045+a047 determine the next actual version.

## Safety conclusion

- Safe edit boundary: dormant exact lineage handoff orchestration.
- High-risk impact: no live capability; derived persistence only.
