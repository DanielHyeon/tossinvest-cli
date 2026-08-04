# Function Logic Map: `checkOrderSize`

- Source: `internal/risk/chain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Intent.Quantity` | positive whole units | sealed strategy intent | `INVALID_ORDER_SIZE` |
| market quote currency | KR/KRW or US/USD | `currencyOf` fixed table | `INPUT_UNAVAILABLE` |
| policy limits | one account-base currency, positive | `Policy.Validate` preflight | `INPUT_UNAVAILABLE` |
| account-base FX | opaque official/identity evidence, exact pair and `Input.Now` | `officialfx.EvidenceAt` via `BindAccountBaseFX` | `INPUT_UNAVAILABLE`; no implicit/caller-raw rate |
| risk/notional capacity | exact rational, whole-unit floor | Guardian policy + frozen FX | refusal before later chain steps |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | quantity parse/non-positive | none | invalid size | legacy quantity tests |
| B3 | unsupported market | none | input unavailable | existing preflight/market tests |
| B4-B7 | missing, stale, wrong-pair or invalid account-base FX; sizing parse/zero | none | input unavailable/invalid size | paired KR identity + US official FX matrix and forged/stale/wrong-pair cases |
| B8 | requested quantity exceeds risk-budget floor | none | invalid size | paired base-risk sizing boundary |
| B9-B10 | max quantity parse/exceeded | none | input unavailable/max-order | existing ceiling tests |
| B11-B13 | base-valued notional failure/exceeds inclusive cap | none | input unavailable/max-order | paired base notional and conservative rounding tests |
| success | every cap admits | none | allowed | paired KR/US table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currencyOf` | bind market to canonical quote currency | pure; unsupported fails closed | CodeGraph + AST |
| account-base sizing helper | convert quote stop width with the same opaque frozen FX and floor once | pure; no retry/fallback | current HEAD + planned paired RED |
| `entryNotional` | exact quote notional | pure; parse error refuses | CodeGraph + AST |
| account-base valuation helper | rate × haircut, then conservative base-unit ceil | opaque FX only; no caller raw rate | a072 currency architecture decision |
| `exceedsInclusiveLimit` | preserve inclusive per-order boundary | pure; currency mismatch refuses | CodeGraph + AST |

## State mutations and fallbacks

- Pure function: no journal, broker, configuration, clock, network or mutation.
- No fallback from US FX failure to quote-currency comparison; KR and US use the same branch shape.

## Safety conclusion

- Safe edit boundary: replace the prior blanket mixed-currency refusal only with an exact opaque-authority conversion; keep quantity and inclusive-cap semantics unchanged.
- High-risk impact: yes. Sizing can raise exposure, so quantity is always floored and monetary reservation/notional is rounded upward.
