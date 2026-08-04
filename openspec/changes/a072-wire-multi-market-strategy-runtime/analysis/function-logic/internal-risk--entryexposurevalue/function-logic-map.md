# Function Logic Map: `EntryExposureValue`

- Source: `internal/risk/chain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| input preflight | current time, valid buy intent/policy/cost model | `preflight` | input unavailable |
| side | BUY only | sealed intent | reductions refused from exposure valuation |
| quote exposure | notional plus cost | shared cost model | input unavailable |
| account-base FX | exact opaque binding at `Input.Now` | `BindAccountBaseFX` | input unavailable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | preflight fails | none | same refusal | existing preflight tests |
| B2 | side is not BUY | none | input unavailable | existing reduction valuation test |
| B3 | quote valuation or sealed FX conversion fails | none | input unavailable | paired authority/rounding tests |
| success | quote exposure converts once and ceils in base units | none | account-base `Money` | paired KR/US table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `preflight` | establish usable entry | pure | CodeGraph + AST |
| `entryNotionalWithCosts` | compute the same quote cash term used by the chain | pure | CodeGraph + AST |
| account-base valuation helper | bind rate×haircut and ceil once | opaque frozen authority only | a072 currency decision |

## State mutations and fallbacks

- Pure exported valuation. Its result must exactly equal the aggregate reservation amount later consumed by the journal; no caller-side second conversion is permitted.

## Safety conclusion

- Safe edit boundary: append the sealed quote→base conversion after the existing quote valuation.
- High-risk impact: yes. Down-rounding a reservation opens risk; conversion is exact rational and upward rounded.
