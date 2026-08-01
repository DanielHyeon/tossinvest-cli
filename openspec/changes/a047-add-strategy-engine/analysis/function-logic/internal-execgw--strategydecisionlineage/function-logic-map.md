# Function Logic Map: `strategyDecisionLineage`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| DecisionRecord | complete 60-field scalar record from opaque decision | strategyengine | JSON encoding error |
| quantity/policy/settings/activation | non-empty exact Guardian and manifest bindings | Guardian/dispatch | journal completeness refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | JSON encoding fails | none | wrapped error | structurally unreachable for current scalar record |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal(record)` | preserve the full DecisionRecord without projection | deterministic scalar JSON | AST + full-payload test |
| SHA-256 | exact payload digest | no fallback | AST + test |

## State mutations and fallbacks

- Returns a value only; journal persistence remains the caller's atomic transaction.
- Projection columns support indexed trace fields while `DecisionPayload` retains all 60 fields exactly.

## Safety conclusion

- Safe edit boundary: pure lineage construction from an already opaque-valid record.
- High-risk impact: yes, loss of a field would launder decision provenance.
