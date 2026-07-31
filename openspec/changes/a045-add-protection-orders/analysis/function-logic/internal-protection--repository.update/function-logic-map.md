# Function Logic Map: `Repository.Update`

- Source: `internal/protection/repository.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Context, expected revision, and proposed fully valid saga. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 invalid saga/revision; B2 SQL error; B3 rows affected error; B4 zero row stale update; else success. | Existing update may rewrite account/profile/market/symbol and jump arbitrary states. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `Saga.Validate`, SQL update; durable local mutation only, no broker call. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- Existing update may rewrite account/profile/market/symbol and jump arbitrary states.

## Safety conclusion

- Safe edit boundary: Transactionally load old row, require immutable identity and an allowed adjacent state transition, then CAS update and commit.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
