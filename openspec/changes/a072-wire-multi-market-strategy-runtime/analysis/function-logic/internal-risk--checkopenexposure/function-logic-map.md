# Function Logic Map: `checkOpenExposure`

- Source: `internal/risk/chain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entry exposure | quote notional plus conservative cost model | `entryNotionalWithCosts` | input unavailable |
| FX | same opaque request-scoped binding used by sizing | `Input.AccountBaseFX` | input unavailable |
| existing exposure | non-negative account-base money | authoritative account snapshot | input unavailable |
| aggregate ceiling | account-base policy money | `Policy.MaxOpenExposure` | boundary reaching limit refuses |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | quote valuation fails | none | typed refusal | cost/input tests |
| B2 | FX/base conversion fails | none | input unavailable | paired missing/stale/wrong-pair tests |
| B3 | existing exposure missing/negative/wrong base currency | none | input unavailable | existing aggregate tests + paired currency test |
| B4 | exact add fails | none | input unavailable | malformed aggregate test |
| B5 | base limit comparison fails | none | input unavailable | policy currency tests |
| B6 | projected base exposure reaches/exceeds cap | none | open-exposure exceeded | paired concurrent-cap arithmetic table |
| success | projected base exposure remains below cap | none | allowed | paired KR/US table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `entryNotionalWithCosts` | compute cash-side quote value once | pure; error refuses | CodeGraph + AST |
| account-base valuation helper | convert and ceil with sealed frozen FX | pure; no fallback/retry | a072 currency decision |
| `magnitudeIn`/`AddDecimal`/`WithinLimit` | validate and compare one account-base aggregate | exact typed failure; ≥ blocks | current HEAD + riskcalc contract |

## State mutations and fallbacks

- No mutation or I/O. The function must not consume quote-currency cash as exposure and must not accept per-market aggregate caps.

## Safety conclusion

- Safe edit boundary: only normalize the new entry term into policy/account base before the existing aggregate add/compare.
- High-risk impact: yes. A lower conversion would fabricate account-wide headroom; conservative ceil is mandatory.
