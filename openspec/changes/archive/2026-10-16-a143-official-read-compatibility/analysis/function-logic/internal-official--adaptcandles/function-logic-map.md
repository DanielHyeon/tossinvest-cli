# Function Logic Map: `adaptCandles`

- Source: `internal/official/candle_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw.Candles` | official wire order: newest to oldest | public REST snapshot | empty remains an empty chart |
| candle timestamp | RFC3339 or absent/invalid | `apiCandle.Timestamp` | preserve existing zero `time.Time` behavior |
| OHLCV | decimal strings | `apiCandle` | preserve existing `parseDecimal` behavior |

Target invariant: only completed general `domain.Chart.Candles` becomes oldest to newest. Symbol, interval, timestamp, decimal conversion, and `FetchedAt` remain unchanged. This function does not consume strict pages and must not convert close time to open time.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range each wire candle | appends one converted candle | no error | multi-candle newest-first fixture becomes chronological after final reversal |
| B2 | timestamp non-empty | parses RFC3339 | parse failure leaves zero time | invalid timestamp stays zero after reversal |
| B3 | parse succeeds | assigns parsed time | no error | exact timestamps in three-candle regression; bad timestamp fixture |
| B4 | left < right in local reversal loop | swaps converted entries | no error | three-candle full-value regression; empty and singleton preserve output |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Parse` | timestamp conversion | ignored parse error preserves zero-time contract | AST B3 |
| `parseDecimal` | price and volume conversion | existing semantics unchanged | AST B1 |
| `time.Now().UTC` | fetch timestamp | unchanged | AST return |

## State mutations and fallbacks

- Allocates and returns a new chart; no I/O, retry, token, setting, journal, or mutation.
- Implemented edit: reverse the newly allocated converted slice after parsing. The raw input slice is not mutated.

## Safety conclusion

- Safe edit boundary: after conversion loop, before chart construction.
- High-risk impact: no. Strict reader and officialbars are separate contracts; regressions protect their ordering/time semantics.
