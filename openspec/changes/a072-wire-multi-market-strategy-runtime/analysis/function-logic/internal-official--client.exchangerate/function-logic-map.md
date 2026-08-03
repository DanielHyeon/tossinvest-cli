# Function Logic Map: `Client.ExchangeRate`

- Source: `internal/official/market_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context/base/quote/client | initialized client and requested currency pair | official Open API endpoint | transport/auth/decode errors propagate; no partial rate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | authenticated GET fails | read-only request only | zero domain rate plus error | exchange-rate integration/error tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Client.get` | perform authenticated `/api/v1/exchange-rate` request | context and classified HTTP errors propagate | CodeGraph + AST |
| `adaptExchangeRate` | preserve exact raw monetary fields | pure conversion | authority-field tests |

## State mutations and fallbacks

- Builds query values locally and performs no client configuration mutation or fallback.

## Safety conclusion

- Safe edit boundary: authoritative callers must hold the surrounding config read lock via `AuthoritativeExchangeRate`.
- High-risk impact: yes, returns raw FX facts later sealed by officialfx.
