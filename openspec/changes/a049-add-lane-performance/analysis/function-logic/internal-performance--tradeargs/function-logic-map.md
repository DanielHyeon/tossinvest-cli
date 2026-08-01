# Function Logic Map: `tradeArgs`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| validated `Trade` | immutable journal-derived value | `Trade.validate` + exact adapter | caller rejects before persistence |
| nullable fields | empty lineage, decision facts, and cost are SQL NULL | explicit missing-state contract | never coerce missing to zero/empty authoritative bytes |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | lineage/optional value empty | none | `nullable` returns SQL NULL | immutable row tests |
| B2 | `CostTotal` empty | none | SQL NULL in `performance_trades.cost_total` | `TestDerivedStorePersistsNullableCostWithoutZeroFabrication` |
| B3 | measured cost present | none | exact decimal string persisted | existing collection tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Lineage.Status` | persist complete vs link_missing | deterministic field completeness | model tests |
| `nullable` / `nullableTime` | bind SQL NULL for absent evidence | no fallback | store tests |

## State mutations and fallbacks

- This function only prepares bound SQL arguments; no string-built query, mutation, poll, or authority exists here.

## Safety conclusion

- Safe edit boundary: make `cost_total` follow the same explicit NULL semantics as other optional evidence.
- High-risk impact: low runtime risk, medium attribution risk pinned by immutable replay tests.
