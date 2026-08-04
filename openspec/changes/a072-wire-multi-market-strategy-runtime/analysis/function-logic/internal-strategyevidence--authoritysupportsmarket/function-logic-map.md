# Function Logic Map: `authoritySupportsMarket`

- Source: `internal/strategyevidence/model.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `authority` | validated official source enum | `authorityValid` | unknown returns false |
| `market` | KR or US | evidence header | a source outside its market returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | OpenDART or KRX | none | true only for KR | market matrix test |
| B2 | SEC EDGAR | none | true only for US | market matrix test |
| B3 | Toss official Open API | none | true for KR and US | market matrix test |
| B4 | unknown authority or market | none | false | market matrix test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure scope classification | no error, timeout or retry path | AST |

## State mutations and fallbacks

- No cross-market substitution: only the explicitly dual-market Toss official source serves both markets.
- There is no fallback from an unsupported source to another authority.

## Safety conclusion

- Safe edit boundary: add one explicit dual-market case; preserve KR-only and US-only cases plus default refusal.
- High-risk impact: yes; this function gates evidence source/market identity.
