# Function Logic Map: `ProjectAcceptedStrategyflowLineage`

- Source: `internal/journal/strategyflow_projection.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed result | verified accepted KR/US strategyflow result | `strategyflow.ProjectAccepted` | reject before journal mutation |
| RiskIntent | exact canonical q_final scope and major-decimal prices | Guardian request | reject drift or invalid decimals |
| metadata | non-empty activation digest and UTC creation instant | paired runtime | reject missing/non-canonical metadata |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | sealed result projection fails | none | wrapped projection error | tampered projection test |
| B2 | RiskIntent is not exact canonical | none | canonical RiskIntent error | risk drift table |
| B3 | activation digest or time invalid | none | metadata error | metadata rejection test |
| B4 | minor-to-major binding fails | none | exact-binding error | paired unit test |
| B5 | valid paired input | none | deterministic v3 lineage | all-six paired test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyflow.ProjectAccepted` | verify opaque result and expose canonical evidence | deterministic; no I/O/retry | AST |
| `exactCanonicalRiskIntent` | enforce exact registered preimage form | deterministic; fail closed | AST |
| `verifyProjectionRiskIntent` | bind q_final and major prices to minor evidence | deterministic; fail closed | pre-edit contract |
| `strategyflowDecisionIdentity` | authenticate canonical v3 payload | deterministic SHA-256 | AST |

## State mutations and fallbacks

- No database, Gateway, lease, broker, activation, or toggle mutation occurs.
- Projection emits v3 only; historical v2 validation remains explicit.

## Safety conclusion

- Safe edit boundary: pure projection before durable authority writes.
- High-risk impact: yes; paired KR/US unit and v2 compatibility tests are mandatory.
