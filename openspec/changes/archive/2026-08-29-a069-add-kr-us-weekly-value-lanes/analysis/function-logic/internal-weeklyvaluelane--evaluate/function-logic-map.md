# Function Logic Map: `evaluate`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| activation/evaluation authority | sealed, plan/market scoped, evaluated_at bound | dormant adapter boundary | LANE_OFF |
| evidence/config/market week | decoded+validated snapshot, exact source, fresh at evaluated_at | a064/canonical calendar ports | typed refusal |
| reservation/leg/cap/risk | same campaign+market scope, active next ordinal, exact cap quantity | a065/a066 sealed snapshots | typed refusal |
| stop/RR | sealed fresh stop; exact checked fixed-point RR | strategy policy | typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | activation authorization absent/invalid | none | LANE_OFF | authorization regression |
| B2 | plan/evidence/source/config invalid | none | typed refusal | isolation/schema tests |
| B3 | invalidation or common exit pending | none | invalidation/refusal | common-exit test |
| B4 | calendar/reservation/leg invalid | none | typed refusal | reservation scope tests |
| B5 | cap stale or quantity mismatch | none | A066_CAP_INVALID | cap freshness/exact test |
| B6 | stop stale/retreat/cap violation | none | stop refusal | sealed-stop tests |
| B7 | RR or risk admission fails | none | typed refusal | RR/risk tests |
| B8 | all gates pass | none | ENTRY_ADD_DECISION | KR/US happy paths |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| evidenceEvaluator | validate sealed immutable PIT snapshot | fail closed | CodeGraph + AST |
| ValidateMarketWeek | exact official week identity/freshness | no fallback | CodeGraph + AST |
| PlannedLegQuantity/RiskCap.validAt | immutable leg quantity and cap | exact bound | CodeGraph + AST |
| effectiveStop/CalculateRR/AdmitRisk | stop, return/risk constraints | checked/refusal | CodeGraph + AST |

## State mutations and fallbacks

- Pure function; no broker/journal/runtime mutation. Every failed authority or evidence gate returns zero quantity.

## Safety conclusion

- Safe edit boundary: package-private sealed inputs and pure output only.
- High-risk impact: yes; sizing and admission decision, covered by fail-closed branch tests.
