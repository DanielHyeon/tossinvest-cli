# Function Logic Map: `Repository.Insert`

- Source: `internal/protection/repository.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| new saga | revision 0/1 normalized to 1 and state exactly PLANNED with no mutation lineage | caller plan + domain validation | invalid saga/transition error, no row |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | normalize revision; reject non-PLANNED/non-1; validate; INSERT | one local DB insert | typed error | planned success and active/registering rejection |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Saga.Validate`, SQL INSERT | persist immutable initial identity only | no retry | CodeGraph + AST |

## State mutations and fallbacks

- Creates one dormant local row; no broker side effect.

## Safety conclusion

- Safe edit boundary: narrow initial state before SQL.
- High-risk impact: yes.
