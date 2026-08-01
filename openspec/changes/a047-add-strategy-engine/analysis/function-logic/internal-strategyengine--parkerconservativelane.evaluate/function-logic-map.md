# Function Logic Map: `ParkerConservativeLane.Evaluate`

- Source: `internal/strategyengine/lane.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed inputs | approved/current candidate, frozen source proof, versioned market bundle, official bar/state/position | opaque package boundaries | ordered typed refusal |
| injected time | evaluation/calendar/indicator time from one sealed bundle | caller clock snapshot | session/candidate/age refusal |
| numeric evidence | exact decimal strings; optional HVN/expansion may be absent | frozen Parker indicator snapshot | indicator or gate-specific refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | approval/scope/source/market/candidate-life invalid | none | candidate/scope/source/indicator refusal | zero/forged and current-life tests |
| B6-B11 | nontrading → open auction → close auction → after-hours → opening skip → cutoff | exact injected times only | exact StockOS source reason | translated regular/early-close and half-open boundary tables |
| B12-B15 | bar clock/state/position proof invalid | none | integrity/session/state/position refusal | zero proof and future-close tests |
| B16-B18 | OHLCV or indicator evidence invalid | exact local parsing only | invalid-bar/illiquid/indicator refusal | opaque-bar invariant, zero-volume, forged indicator tests |
| B19-B27 | VWAP→slope→EMA9→bullish→LVN→tangled→optional expansion gates | exact local arithmetic only | first matching source refusal | frozen thresholds, EMA exact edge, fake-breakout precedence |
| B28-B32 | RR→optional HVN→age fails | exact local arithmetic only | first matching refusal | structural RR floor, HVN, nanosecond age tests |
| B33-B36 | live price presence/parse/absolute drift | exact local arithmetic only | drift refusal | missing/malformed/nonpositive/positive-negative edge tests |
| B37-B38 | canonical identity and final mint | pure JSON/SHA and exact revalidation | decision refusal | fixed-schema identity invariant and direct mint mutation table |
| Success | every gate passes | none outside returned value | opaque Decision | synthetic derivation plus translated StockOS final-bar/indicator arithmetic |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| exact decimal helpers | avoid float/asserted derived values | pure; invalid values fail closed | CodeGraph + AST |
| `decisionIdentity` / `mintDecision` | bind and revalidate complete evidence | pure; error becomes decision refusal | CodeGraph + AST |

## State mutations and fallbacks

- No state mutation, I/O, retry, or fallback. Evaluation returns one immutable decision or one ordered refusal.

## Safety conclusion

- Safe edit boundary: source order is preserved; RR, age, drift and target are derived internally.
- High-risk impact: yes — activation decision boundary, fully covered by parity and nanosecond edge tests.
