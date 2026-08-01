# Function Logic Map: `Store.PruneDue`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| now | explicit UTC maintenance instant | scheduler/caller | invalid cadence state fails closed |
| retention/cadence | 90d cutoff, 24h only after backlog drained | OpenSpec constants | overdue backlog never advances cadence marker |
| batch bounds | <=500 rows/transaction and finite batches/run | OpenSpec + implementation constant | remaining backlog is reported for immediate reschedule |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | loop below max four batches | per-batch immediate transaction | bounded continuation | backlog test |
| B2 | BeginTx fails | none | error | DB error contract |
| B3 | first batch | reads cadence | continue | cadence tests |
| B4 | cadence SELECT errors other than no-row | rollback | error | DB error contract |
| B5 | cadence value exists | parses RFC3339Nano | continue or B6 error/B7 skip | cadence tests |
| B6 | cadence parse fails | rollback | error | invalid metadata contract |
| B7 | now is before last+24h | commits read-only tx | skipped | cadence test |
| B8 | read-only commit fails | none durable | error | DB error contract |
| B9 | bounded DELETE fails | rollback | error | DB error contract |
| B10 | RowsAffected fails | rollback | error | driver contract |
| B11 | backlog EXISTS fails | rollback | error | DB error contract |
| B12 | no backlog remains | writes completed cadence in same tx | continue | drain test |
| B13 | cadence write fails | rollback | error | DB error contract |
| B14 | prune commit fails | rollback/recovery | error | crash/transaction contract |
| B15 | batch deleted count raises max | updates result only | continue | max-batch test |
| B16 | batch lock duration raises max | updates result only | continue | lock-duration test |
| B17 | backlog drained | returns complete; otherwise next bounded batch/reschedule | result | drain/backlog/influx tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| cadence SELECT | checks last completed drain | read-only | current HEAD + AST |
| bounded DELETE | oldest `(observed_at,id)` using covering index | max 500, one transaction | EXPLAIN/prune tests |
| backlog EXISTS | detects remaining rows before publishing cadence | same tx/cutoff | prune tests |

## State mutations and fallbacks

- Deletes only rebuildable `price_observations`; authoritative trades, lineage, snapshots and journal stay intact.
- Total work per call is bounded; each writer lock duration is measured separately.

## Safety conclusion

- Safe edit boundary: derived raw retention maintenance.
- High-risk impact: derived data deletion; bounded/recoverable and never touches journal.
