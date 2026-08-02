# Function Logic Map: `ReconcileDriver.adopt`

- Source: `internal/app/engine/adoption.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidates | folded holdings already selected by the adoption gate | `ReconcileDriver.Run` | empty set returns without a broker read |
| quote observations | market-qualified, positive, currency-validated values | `observeCandidates` | absent/error observations increment Deferred and never reach `adoptOne` |
| readAt/staleness | request-start timestamp and positive configured/default bound | engine clock + `PriceStaleness` | a stale first observation defers the remaining batch |
| cycle | non-nil current reconciliation counters/error | reconciliation driver | records failure/deferred/adopted truth; does not hide unmanaged exposure |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 175 | `if len(candidates) == 0 {` | initializes empty adopted set only | returns without broker/journal access | `TestAnIncludedSymbolIsAdoptedWithTheSwitchOff` regression path |
| B2 | `if` at line 180 | `if err != nil {` | increments Deferred for the full batch | returns empty adopted set | broker-read failure coverage in reconciliation suite |
| B3 | `if` at line 182 | `if cycle.Err == nil {` | preserves the first cycle error | does not overwrite an earlier error | broker-read failure coverage in reconciliation suite |
| B4 | `if` at line 189 | `if bound <= 0 {` | substitutes conservative default staleness | continues | `TestAStalePriceDefersTheAdoption` |
| B5 | `range` at line 192 | `for _, c := range candidates {` | processes each folded candidate independently | returns the IDs actually adopted | `TestAnIncludedSymbolIsAdoptedWithTheSwitchOff`; `TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0` |
| B6 | `if` at line 195 | `if !ok {` | increments Deferred; no journal write | continues to next candidate | `TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency` |
| B7 | `if` at line 201 | `if age := d.clk.Now().Sub(readAt); age > bound {` | defers the unadopted remainder and logs reason | returns before any later candidate can persist stale t0 | `TestAStalePriceDefersTheAdoption` |
| B8 | `if` at line 210 | `if d.adoptOne(ctx, c, observed) {` | success records ID and increments Adopted; failure increments Deferred | continues; success cannot be inferred from quote alone | `TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0`; adoption failure regressions |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `len` | execute the explicit dependency at line 175 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `d.observeCandidates` | perform the single pre-transaction batched quote read | may use the configured query retrier; error defers the full batch | B2-B3 + quote identity tests |
| `adoptionQuoteKey` | execute the explicit dependency at line 193 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `Sub` | execute the explicit dependency at line 201 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `d.clk.Now` | execute the explicit dependency at line 201 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `d.logDeferred` | execute the explicit dependency at line 206 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `fmt.Sprintf` | execute the explicit dependency at line 206 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `d.adoptOne` | persist one adoption and open its exit state only after validated quote/freshness | returns false on any derivation/journal failure; caller increments Deferred | B8 + adoption lifecycle tests |

## State mutations and fallbacks

- `cycle.Deferred`, `cycle.Err`, `cycle.Adopted`, and the local adopted-ID set are the only direct mutations.
- The only persistence call is `adoptOne`, reached after quote identity and age pass; a missing/wrong-currency/ambiguous/stale quote cannot freeze t0 or a synthetic stop.
- This path creates protection state but places no order and does not resolve reconciliation.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
