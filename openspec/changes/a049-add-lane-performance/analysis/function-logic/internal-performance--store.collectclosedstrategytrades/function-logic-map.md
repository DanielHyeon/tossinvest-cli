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
| B1 | reader nil | none | error | handoff test |
| B2 | reader error | none | wrapped error | handoff error test |
| B3 | trade exact replay | no duplicate rows | snapshot included in result | restart/replay test |
| B4 | trade divergence/append failure | prior per-trade commits remain; current transaction rolls back | partial result + error | replay divergence test |

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
