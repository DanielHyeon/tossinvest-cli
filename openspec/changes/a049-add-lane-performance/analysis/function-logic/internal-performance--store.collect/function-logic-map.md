# Function Logic Map: `Store.Collect`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| trade | validated immutable journal-derived trade | exact lineage reader/caller | invalid or divergent persisted bytes fail closed |
| observations | exact position, caller-owned values | caller; no read/poll capability | mismatch/divergence aborts transaction |
| calculatedAt | deterministic snapshot identity component | caller | zero is rejected through snapshot validation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | trade validation fails | none | error | validation tests |
| B2 | iterate observations | none | continue | collection tests |
| B3 | observation validation fails | none | error | validation tests |
| B4 | observation position differs | none | error | foreign-position contract |
| B5 | calculatedAt zero | none | error | snapshot identity validation |
| B6 | BeginTx fails | none | wrapped error | context/DB error contract |
| B7 | trade compare-and-append fails | pending tx rolls back | conflict/error | divergence tests |
| B8 | observation compare-and-append fails | pending tx rolls back | conflict/error | divergence tests |
| B9 | snapshot compare-and-append fails | pending tx rolls back | conflict/error | divergence tests |
| B10 | commit fails/process dies | transaction rolls back | error; otherwise snapshot success | phase crash/replay tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Measure` | derives side-aware metrics from supplied observations | pure/no poll | AST + model tests |
| compare-and-append helpers | enforce immutable equality | divergence aborts whole transaction | store tests |
| `Commit` | publishes trade, rows and snapshot together | no retry or partial success | crash tests |

## State mutations and fallbacks

- The transaction is append-only. Exact replay is deliberately accepted; divergent bytes are never updated.
- No symbol/time lineage guess, external quote read, broker mutation, config write, or LIVE capability exists.

## Safety conclusion

- Safe edit boundary: rebuildable derived read-model collection.
- High-risk impact: persistence/reporting integrity; no trading side effect.
