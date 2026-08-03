# Function Logic Map: `Console.history`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | live request context | HTTP handler | journal/name reads terminate with request |
| `c.opts.JournalPath` | empty, missing, compatible read-only SQLite | `openJournal` | return typed journal state with no rows |
| account refs | zero or more account identities | `ReadOnly.AccountRefs` | mark journal failed and return without partial account mixing |
| trips/events | frozen journal projections | account-scoped read-only queries | mark journal failed and return; never recompute |
| instrument metadata | zero or more display names for unique market+symbol refs | optional narrow official reader + bounded cache | preserve every symbol and row; expose metadata warning |
| metadata admission | optional cross-process lease in command adapter | `ratebudget` + profile run marker | a run intent/active lease yields cached or code-only history |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal cannot be opened/read | only `v.Journal` display state | early return with no rows | `TestTheHistoryScreenIsHonestWhenNothingHasClosed` |
| B2 | account refs query fails | set journal failed detail | early return; no invented empty state | journal failure harness coverage |
| B3 | iterate account refs | append only request-local rows | continue until a scoped query fails | account-view/history tests |
| B4 | trips query fails for current account | set journal failed detail | early return; no mixed partial result | read-only journal query coverage |
| B5 | iterate frozen trips | append immutable display values | continue | `TestTheHistoryScreenShowsTheFrozenRoundTrip` |
| B6 | events query fails for current account | set journal failed detail | early return; no mixed partial result | read-only journal query coverage |
| B7 | event count equals `exitEventWindow` | set `Truncated` | continue and render explicit limit notice | `TestTheHistoryScreenStatesItsOwnLimit` |
| B8 | iterate exit events | append immutable display values | continue | `TestTheHistoryScreenStatesItsOwnLimit` |
| B9 | account produced trips or events | append masked account | continue; multi-account column derives from count | history multi-account coverage |
| B10 | iterate completed trips after lookup | attach request-local safe name by market+symbol | continue; missing key leaves empty name | `TestA061HistoryShowsCodeAndNameForTripsAndEvents` |
| B11 | iterate exit events after lookup | attach request-local safe name by market+symbol | continue; missing key leaves empty name | `TestA061HistoryShowsCodeAndNameForTripsAndEvents` |
| Metadata fallback | reader absent/fails/name missing | attach no name; record display warning on failure/hold | symbol and frozen facts remain visible | `TestA061HistoryNameFailureKeepsFrozenRowsAndSymbols` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openJournal` | open exact engine journal read-only | typed state, no migration/write/fallback journal | AST + read-only invariant tests |
| `AccountRefs` | enumerate account partitions | first error stops projection | AST B2 |
| `AccountTradeTrips` | fetch frozen outcomes plus symbol/entry join | first error stops projection | AST B3 + journal account-view tests |
| `AccountExitEvents` | fetch newest bounded judgement window | first error stops projection | AST B4/B5 + journal account-view tests |
| `attest.Mask` | prevent account identifier disclosure | pure display transform | AST per-account loop |
| `sort.SliceStable` | chronological display across accounts | in-memory request-local mutation only | AST tail |
| `instrumentNameCache.get` | batch descriptive metadata for unique refs | optional; 10s total timeout, bounded cache/lookup, verification hold; failure is label-only | a061 regression and adapter tests |
| `verifyHold` | fast-path existing live verification state before optional reader | cached/code-only fallback | hold and POST regression tests |

## State mutations and fallbacks

- `historyView` is request-local. No journal, config, account, or cache write occurs.
- Any journal query failure replaces the page's journal state with explicit failure and returns.
- Metadata enrichment is deliberately weaker than journal truth: failure keeps the symbol and all frozen rows.
- Cross-process admission is owned by the command adapter; verification intent is published before its exclusive lease wait, closing the check/use race.
- Sorting changes only in-memory row order after all account reads finish.

## Safety conclusion

- Safe edit boundary: request-local name decoration runs after journal rows and sorting are complete; journal queries, outcome values, ordering, and mutation surfaces are unchanged.
- High-risk impact: no. The function is read-only, but the new external read must remain narrow, bounded, escaped, and fail-soft.
