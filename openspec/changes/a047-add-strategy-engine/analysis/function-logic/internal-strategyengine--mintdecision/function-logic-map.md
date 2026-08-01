# Function Logic Map: `mintDecision`

- Source: `internal/strategyengine/decision.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance | exact lane/source/config/calendar/indicator/bar/candidate bindings | lane-created record | error |
| clocks | candidate `[last,validUntil)`, signal `[close,close+15s+1ns)`, fresh state/position | injected evaluation | error |
| arithmetic | stop 0.7%, target 3R, RR and drift recomputed from raw observations | frozen Parker constants | error |
| identity | SHA-256 of full record excluding identity field | canonical JSON | error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | provenance mismatch | none | error | zero/forged and identity tests |
| B2 | clock/session/freshness mismatch | none | error | session/age/candidate tests |
| B3 | decimal/optional evidence invalid | exact parse only | error | gate tests |
| B4 | recomputed prices/RR/drift/HVN/reasons differ | exact arithmetic only | error | golden parity test |
| B5 | identity differs | SHA only | error | decision identity validation |
| B6 | all equal | none | valid opaque Decision | golden test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decisionDecimal` | exact evidence parsing | pure, no fallback | AST |
| `decisionIdentity` | canonical evidence identity | pure SHA-256 | AST |

## State mutations and fallbacks

- No external mutation or fallback; the record is copied into the returned opaque value.

## Safety conclusion

- Safe edit boundary: independently recomputes every caller-sensitive derived field.
- High-risk impact: yes — last minting boundary before dispatch.
