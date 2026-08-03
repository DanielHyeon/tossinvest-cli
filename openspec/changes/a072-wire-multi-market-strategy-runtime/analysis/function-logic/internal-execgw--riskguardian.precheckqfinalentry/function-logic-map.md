# Function Logic Map: `RiskGuardian.PrecheckQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request + Guardian | exact policy/digest, canonical market/currency/symbol, positive candidate, exact owner, opaque FX current at Guardian clock | RiskGuardian state + opaque officialfx authority | typed q_final refusal; no journal/broker call |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | nil/stale/scope/quantity/Guardian-currency/owner/reserve-currency failures | none | typed refusal | existing q_final tests |
| B9 | public FX DTO/source-label substitution or missing opaque authority | none | currency unresolved | new authority bypass tests |
| B10-B14 | Guardian cap, admission or ordinary Guardian chain failure | none | typed refusal | existing focused tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| opaque FX ReserveAt | verify seal, exact pair and freshness at Guardian clock | no fallback | new officialfx authority tests |
| risk.StrategyEntryQuantity | existing Guardian cap | exact decimal/error | existing tests |
| riskbucket.CalculateAdmission | five-bucket q_final | pure typed refusal | current source |
| evaluateChain/EntryExposureValue | existing Guardian safety chain | fail closed | current source |

## State mutations and fallbacks

- Deep-copies caller slices, overwrites QCandidate, EvaluatedAt, QExistingGuardian and FX DTO from opaque authority.

## Safety conclusion

- Safe edit boundary: caller FX fields are never authority and cannot survive overwrite.
- High-risk impact: yes — Guardian sizing; no mutation in this phase.
