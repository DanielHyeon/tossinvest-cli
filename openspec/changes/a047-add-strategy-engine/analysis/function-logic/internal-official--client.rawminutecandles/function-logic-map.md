# Function Logic Map: `Client.RawMinuteCandles`

- Source: `internal/official/candle_raw.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| client | non-nil | official client | local error |
| request identity | market `KR`, canonical six-digit symbol | caller request | local error before network |
| query | fixed `1m`; explicit adjusted flag; optional count/cursor | official endpoint contract | transport/API error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 48: `if c == nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 51: `if market != "KR" \|\| !krxRawCandleSymbol.MatchString(symbol) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 57: `if count > 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 60: `if before != "" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 65: `if err := c.get(ctx, "/api/v1/candles", q, &raw); err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `range` at source line 74: `for _, candle := range raw.Candles {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Client.get` | official authenticated GET | propagates auth/HTTP/decode error; no retry added here | CodeGraph + AST |

## State mutations and fallbacks

- Read-only API call. No fallback to WTS; the returned source is minted only after successful official decoding.

## Safety conclusion

- Safe edit boundary: validation precedes transport, DTO strings remain unchanged.
- High-risk impact: yes — this provenance is required before a verified strategy bar can be minted.
