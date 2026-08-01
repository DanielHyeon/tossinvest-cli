# Function Logic Map: `appendSnapshot`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| tradeID/snapshot | validated deterministic measurement identity and metric bytes | `Collect`/`Measure` | missing parent or divergent replay aborts transaction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | metric canonicalization fails | none | error | snapshot validation paths |
| B2 | snapshot identity exists | none | compare header/metrics | replay tests |
| B3 | existing lineage status differs | none | `ErrImmutableConflict` | divergent replay test |
| B4 | identity lookup DB error other than no-row | none | wrapped error | DB error contract |
| B5 | new snapshot INSERT fails | pending tx only | error | crash/all-or-none tests |
| B6 | LastInsertId fails | pending tx only | error | driver contract |
| B7 | iterate six canonical records | inserts pending metric rows | continue | collect test |
| B8 | metric INSERT fails | pending tx only | error | crash/all-or-none tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| snapshot lookup | finds unique trade/calculatedAt/semantics identity | no row inserts; DB error fails | current HEAD + AST |
| metric comparison/INSERT | preserves all value, gross, cost, observation and provenance fields | exact equality required | store tests |

## State mutations and fallbacks

- Append-only measurement history; no UPDATE or newest-value replacement.

## Safety conclusion

- Safe edit boundary: derived snapshot persistence.
- High-risk impact: reporting integrity only.
