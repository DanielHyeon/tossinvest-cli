# Function Logic Map: `signalsVetoTallyFrom`

- Source: `internal/console/signals.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidate tally | totals, code maps, reasons | `candidate.VetoTally` | unexpected pass is explicitly labelled and anomalies retained |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | passed non-zero in dormant mode | local note change | continue | signals tally tests |
| B2 | iterate anomalies | local alarm append | continue | alarm tests |
| B3 | iterate copied D3 codes | local count append | continue | signals tests |
| B4 | iterate reasons | local map fill | sorted render | signals tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Anomalies`, `OrderedVetoCodes`, sorting helpers | deterministic read model | pure | CodeGraph + AST |

## State mutations and fallbacks

- Fresh console DTO only. No threshold or candidate state is changed.

## Safety conclusion

- Safe edit boundary: immutable D3 order projection.
- High-risk impact: no; render-only.
