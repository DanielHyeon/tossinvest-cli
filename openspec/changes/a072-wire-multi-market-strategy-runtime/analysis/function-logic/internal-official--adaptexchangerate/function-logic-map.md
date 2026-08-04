# Function Logic Map: `adaptExchangeRate`

- Source: `internal/official/market_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw.BaseCurrency` / `raw.QuoteCurrency` | Official response strings | `/api/v1/exchange-rate` response | Adapter preserves both exact strings and also derives legacy `Code`; it performs no authority validation |
| `raw.Rate` / `raw.MidRate` | Official decimal strings | Same response | `parseDecimal` is display-only and maps empty/invalid input to float zero |
| `raw.ValidFrom` / `raw.ValidUntil` | Official datetime strings | Same response | Adapter preserves both exact strings; parsing is deferred to the fail-closed evidence boundary |
| returned `domain.ExchangeRate` | Read-only transport value | `adaptExchangeRate` | Preserves every authority-bearing raw field without changing legacy display floats |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (branchless happy path) | The function is one unconditional struct projection | None | Always returns a value; authoritative validation belongs to the sealed pure adapter | Unit test pins exact raw strings and legacy display floats |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseDecimal` | Preserve legacy `Base`/`Close` display compatibility | No error return; empty/invalid becomes zero and therefore must never be an authority source | CodeGraph + pre-edit AST |

## State mutations and fallbacks

- No mutation, I/O, retry, timeout, clock read, journal write, broker call or configuration binding.
- The legacy float-zero fallback remains display-only. The new official FX authority adapter consumes
  only the raw decimal/time fields and fails closed independently.

## Safety conclusion

- Safe edit boundary: add lossless field copies to the returned domain value; do not add validation,
  clocks, policy or capability to this transport adapter.
- High-risk impact: low for this function, but its new raw fields feed a high-risk monetary authority
  boundary. Tests must prove raw preservation and prove that invalid raw input cannot mint FX evidence.
