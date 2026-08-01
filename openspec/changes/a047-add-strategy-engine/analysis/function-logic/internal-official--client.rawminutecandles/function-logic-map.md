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
| B1 | nil client | none | error | nil boundary |
| B2 | unsupported/non-canonical identity | none | error before HTTP | unsupported market test |
| B3 | optional count/before present | query only | continue | query assertions |
| B4 | official read fails | token/HTTP read only | propagated error | client transport suite |
| B5 | official read succeeds | none | lossless page with identity/interval/adjusted/source | preservation test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Client.get` | official authenticated GET | propagates auth/HTTP/decode error; no retry added here | CodeGraph + AST |

## State mutations and fallbacks

- Read-only API call. No fallback to WTS; the returned source is minted only after successful official decoding.

## Safety conclusion

- Safe edit boundary: validation precedes transport, DTO strings remain unchanged.
- High-risk impact: yes — this provenance is required before a verified strategy bar can be minted.
