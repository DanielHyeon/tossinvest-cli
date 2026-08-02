# Function Logic Map: `ReconcileDriver.observeCandidates`

- Source: `internal/app/engine/adoption.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidates | folded positions with non-empty market and symbol | reconciliation fold output | malformed identity omitted before the quote request |
| quote identity | positive last plus a symbol belonging to exactly one candidate market | official price response + candidate index | zero, unrelated, or cross-market ambiguous quote omitted |
| expected currency | KR→KRW, US→USD after trim/case normalization | `adoptionCurrency` | empty/mismatch/unknown-market quote omitted |
| readAt | clock value captured before the broker call | engine clock | returned only with successful batch; conservative age includes retry latency |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range` at line 225 | `for _, c := range candidates {` | builds normalized symbol→market set and deduplicated request | continues | `TestObserveCandidatesRequiresMarketCurrencyIdentity`; `TestObserveCandidatesRefusesCrossMarketDuplicateSymbol` |
| B2 | `if` at line 228 | `if symbol == "" \|\| market == "" {` | none | skips malformed identity | quote-identity table plus folded-input regression |
| B3 | `if` at line 232 | `if markets == nil {` | creates market set and appends symbol exactly once | continues | `TestObserveCandidatesRefusesCrossMarketDuplicateSymbol` |
| B4 | `if` at line 257 | `if d.opts.Retrier != nil {` | performs budgeted official quote query | continues with result/error | retrier reconciliation tests |
| B5 | `else` at line 259 | `} else {` | performs one direct official quote query | continues with result/error | quote-identity tests |
| B6 | `if` at line 262 | `if err != nil {` | none | returns wrapped error and zero timestamp | broker-read error regression |
| B7 | `range` at line 268 | `for _, q := range quotes {` | builds validated observation map only | returns map/readAt | all quote-identity tests |
| B8 | `if` at line 269 | `if q.Last <= 0 {` | none | skips non-positive quote | zero-price reconciliation regression |
| B9 | `if` at line 276 | `if len(markets) != 1 {` | none | skips unrelated or cross-market ambiguous symbol | `TestObserveCandidatesRefusesCrossMarketDuplicateSymbol` |
| B10 | `range` at line 282 | `for candidateMarket := range markets {` | selects the sole proven market | continues | `TestObserveCandidatesRequiresMarketCurrencyIdentity` |
| B11 | `if` at line 286 | `if !ok \|\| !strings.EqualFold(strings.TrimSpace(q.Currency), expectedCurrency) {` | none | skips unknown-market, blank, or wrong-currency quote | `TestObserveCandidatesRequiresMarketCurrencyIdentity`; `TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `make` | execute the explicit dependency at line 224 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `len` | execute the explicit dependency at line 224 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.ToUpper` | execute the explicit dependency at line 226 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.TrimSpace` | execute the explicit dependency at line 226 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.ToLower` | execute the explicit dependency at line 227 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `append` | execute the explicit dependency at line 235 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `sort.Strings` | execute the explicit dependency at line 239 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `d.clk.Now` | execute the explicit dependency at line 248 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `d.opts.Prices.Prices` | one batched official quote read for deduplicated symbols | error propagates through B6; no persistence occurs here | AST + quote fixtures |
| `d.opts.Retrier.Query` | apply the existing price query retry budget | retry latency is included because readAt is stamped before the call | B4 + existing retrier tests |
| `read` | execute the explicit dependency at line 260 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `fmt.Errorf` | execute the explicit dependency at line 263 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `adoptionCurrency` | resolve the only allowed currency for a candidate market | unknown market returns false and is omitted | B11 + currency table test |
| `strings.EqualFold` | execute the explicit dependency at line 286 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `adoptionQuoteKey` | execute the explicit dependency at line 289 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `decimalOf` | execute the explicit dependency at line 289 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- Mutations are confined to local request/index/result collections and the broker response slice.
- No journal or order call occurs here. Duplicate cross-market symbols are judged before any caller can persist adoption/t0.
- Failure, ambiguity, non-positive prices, unknown markets, and currency mismatch all omit the observation, causing later adoption to defer fail-closed.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
