# Function Logic Map: `Evidence.EvidenceAt`

- Source: `internal/officialfx/evidence.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed evidence/evaluation time | complete seal and current source+policy window | opaque officialfx evidence | typed invalid/stale refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | zero/tampered seal | none | ErrInvalidEvidence | tamper tests |
| B2 | time outside intersection window | none | ErrEvidenceNotCurrent | boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sealFor | recompute immutable preimage | pure | current source |

## State mutations and fallbacks

- Refactor returns opaque reserve facts instead of a public riskbucket.FXEvidence; downstream must retain this Evidence.

## Safety conclusion

- Safe edit boundary: getters expose values only after validation and cannot become authority by themselves.
- High-risk impact: yes — last freshness boundary before q_final.
