# Function Logic Map: `mintDecision`

- Source: `internal/strategyengine/decision.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance | exact lane/source/config/calendar/indicator/bar/candidate bindings | lane-created record | error |
| clocks | candidate `[last,validUntil)`, signal `[close,close+15s+1ns)`, fresh state/position, cutoff exactly `session close - 45m` | injected evaluation + frozen StockOS config | error |
| arithmetic | stop 0.7%, target 3R, RR and drift recomputed from raw observations | frozen Parker constants | error |
| identity | SHA-256 of full record excluding identity field | canonical JSON | error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | provenance mismatch | none | error | zero/forged and identity tests |
| B2 | clock/session/freshness mismatch or caller-selected cutoff | none | error | session/cutoff/age/candidate tests |
| B3-B4 | required positive evidence iteration/validation fails | exact parse only | error | first/last forged positive fields |
| B5 | tangled evidence malformed or below threshold | exact parse only | error | direct forged tangled row |
| B6-B8 | optional expansion/HVN iteration, presence, or parsing fails | exact parse only | error | first/second forged optional fields plus absent success |
| B9 | live-entry delta is negative | normalize local rational to absolute value | continue | direct negative-drift valid remint |
| B10 | recomputed close/stop/target/RR/drift differs | exact arithmetic only | error | direct forged derived-price row |
| B11 | unobserved live price differs from close | none | error | direct forged and valid fallback rows |
| B12-B13 | HVN is present and below LVN forward space | exact parse/compare only | error | direct forged HVN plus absent success |
| B14 | ordered accept reasons differ | none | error | direct forged reason-order row |
| B15 | canonical identity calculation fails or differs | SHA only | error | direct forged identity row |
| Success | all bindings equal | none | valid opaque Decision | base remint and negative/optional success tests |

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
