# Function Logic Map: `evaluateEvidence`

- Source: `internal/weeklyvaluelane/schema.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| evidence | decoder-sealed immutable snapshot with bounded identities | strict decoder | schema refusal |
| config | package-sealed exact market/source/schema/model | config owner | config refusal |
| PIT/revision/dilution/vector | complete, ordered, canonical and in digest preimage | a064 adapter | typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | config or evidence seal invalid | none | config/schema refusal | literal/tamper test |
| B2 | market/source/schema/identity/revision invalid | none | typed refusal | schema tests |
| B3 | PIT/freshness/dilution invalid | none | typed refusal | PIT tests |
| B4 | unit/vector/arithmetic invalid | none | typed refusal | arithmetic tests |
| B5 | canonical digest mismatch | none | schema refusal | digest test |
| B6 | exact replay valid | none | accepted+decision digest | replay test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| evidenceSnapshotDigest | canonical full immutable preimage | exact SHA-256 | CodeGraph + AST |
| evidenceDecisionDigest | snapshot+model decision preimage | exact SHA-256 | CodeGraph + AST |

## State mutations and fallbacks

- No mutation; caller literals and post-decode field mutation invalidate private seal.

## Safety conclusion

- Safe edit boundary: strict schema/PIT validation only.
- High-risk impact: yes; prevents future-data and revision substitution.
