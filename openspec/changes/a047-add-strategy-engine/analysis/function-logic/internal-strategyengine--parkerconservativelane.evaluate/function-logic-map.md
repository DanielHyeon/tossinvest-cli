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
| B1 | approval/scope/source invalid | none | candidate/scope/source refusal | zero/forged proof tests |
| B2 | candidate future/stale or calendar session invalid | none | candidate/session refusal | activation/current-life/session boundary tests |
| B3 | bar/state/position invalid | none | integrity/state/position refusal | proof table |
| B4 | malformed bar/indicator | exact local parsing only | invalid-bar/indicator refusal | malformed input tests |
| B5 | source gates 8a-8f fail | exact local arithmetic only | VWAP/slope/EMA9/LVN/tangled/expansion refusal | frozen gate table |
| B6 | RR→HVN→age→drift fails | exact local arithmetic only | first matching refusal | derived-boundary tests |
| B7 | every gate passes | none outside returned value | opaque Decision | synthetic derivation plus translated StockOS final-bar/indicator parity tests |

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
