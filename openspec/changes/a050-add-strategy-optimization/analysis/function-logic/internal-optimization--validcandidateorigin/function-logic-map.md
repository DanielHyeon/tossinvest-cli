# Function Logic Map: `validCandidateOrigin`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidate source and reason | only server/evidence preset or rollback pairs | persisted candidate | false causes fail-closed apply |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | server preset | none | matching reason result | metadata tamper test |
| B2 | evidence candidate | none | matching reason result | metadata tamper test |
| B3 | rollback | none | matching reason result | rollback test |
| B4 | unknown source | none | false | metadata tamper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| source switch | validates immutable origin tuple | no I/O | AST |

## State mutations and fallbacks

- Pure candidate provenance validation.

## Safety conclusion

- Safe edit boundary: candidate integrity.
- High-risk impact: no.
