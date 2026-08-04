# Function Logic Map: `RiskGuardian.PrecheckQFinalCampaignFirstLeg`

- Source: `internal/execgw/riskguardian_first_leg.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `request.Entry` | valid sealed KR/US q_candidate request with no caller prospective token | `RiskGuardian.PrecheckQFinalEntry` | typed refusal; no collection/write |
| `request.Result` | accepted/pure sealed `strategyflow.Result`, candidate quantity equal to entry q_candidate | `strategyflow.Result` + journal projector | typed refusal; no write |
| execution prices | KRW scale-0 identity or USD scale-2 canonical major decimals | sealed price provenance + Guardian entry intent | conversion failure or mismatch refuses before journal |
| activation/attempt metadata | non-empty digest/attempt, revision >= 1 | engine-owned authority loader | typed refusal; no write |
| campaign CAS | exact campaign id, current generation/version; prospective lineage generation = current+1 | journal campaign state | typed refusal; no write |
| weekly reservation | exact opaque schema-v27 market/week binding for KR/US weekly lanes; nil for every non-weekly lane | production proposal authority + journal read-only reservation | typed refusal; no write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | q_final entry precheck refuses | none | propagate refusal | paired KR/US q_final tests |
| B2 | q_candidate differs from sealed result quantity | none | risk-calculation refusal | downsizing/tamper table |
| B3 | result scope, lane, campaign or canonical major execution prices differ | read-only projector only | risk-calculation refusal | paired tamper/unit table |
| B4 | current generation overflows or prospective != current+1 | none | owner-conflict refusal | flat 0/0 and generation tamper tests |
| B5 | metadata incomplete | none | risk-calculation refusal | metadata tamper tests |
| B6 | all bindings exact, including q_candidate 20 / q_final 10 | read-only projection | opaque precheck | paired six-lane downsizing test |
| B7 | weekly binding is absent, cross-market, divergent, or smuggled onto a non-weekly lane | none | weekly-authority refusal | paired weekly/non-weekly binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PrecheckQFinalEntry` | seal Guardian q_final without mutation | immediate error; no retry | CodeGraph + AST |
| `journal.QFinalPolicyVersion` | derive Guardian-owned q_final policy | fail closed | CodeGraph + AST |
| `journal.ProjectAcceptedStrategyflowLineage` | bind sealed result to Guardian RiskIntent | read-only; fail closed | CodeGraph + AST |
| `PriceProvenance.MajorDecimal` | convert authenticated minor evidence without floats | deterministic; fail closed | paired unit contract |

## State mutations and fallbacks

- No journal, Gateway or broker mutation. The returned type is opaque and keeps the exact sealed entry, projected strategy, authoritative campaign CAS and optional exact weekly reservation binding.

## Safety conclusion

- Safe edit boundary: accept only sealed strategyflow evidence plus authority metadata; never accept a caller-built `StrategyPlanRequest`.
- High-risk impact: yes — sizing/Guardian/lineage admission; paired fail-closed tests required.
