# Function Logic Map: `VetoTally.Anomalies`

- Source: `internal/candidate/tallycheck.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| tally counters/maps | non-negative counts by D3 code | `TallyVetoes` or persisted read model | contradictions become ordered findings |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied D3 order | local findings | continue | tally alarm tests |
| B2 | passed+raised+missing exceeds total | append overcount anomaly | continue | tally alarm tests |
| B3 | threshold-absent and passed both non-zero | append fail-closed anomaly | return findings | tally alarm tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes` | deterministic contradiction ordering | pure | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only a local findings slice; does not repair or hide invalid tallies.

## Safety conclusion

- Safe edit boundary: immutable ordering accessor only.
- High-risk impact: no; detector semantics unchanged.
