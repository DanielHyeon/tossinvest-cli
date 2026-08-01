# Function Logic Map: `validateMarketCalendarDate`

- Source: `internal/official/calendar_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| date | empty or exact ten-byte Gregorian `YYYY-MM-DD` | official calendar query contract | invalid returns error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | empty date | none | nil (server default) | exact-date/read tests |
| B2 | length or hyphen positions differ | none | validation error | exact-date test |
| B3 | parse fails or format does not round-trip | none | validation error | impossible-date test |
| success | exact real date | none | nil | typed calendar fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Parse` + round trip | Gregorian semantic and canonical lexical check | pure | AST + tests |

## State mutations and fallbacks

- Pure pre-network validation shared by typed and legacy reads.

## Safety conclusion

- Safe edit boundary: calendar query date only.
- High-risk impact: medium; wrong date would corrupt entry evidence.
