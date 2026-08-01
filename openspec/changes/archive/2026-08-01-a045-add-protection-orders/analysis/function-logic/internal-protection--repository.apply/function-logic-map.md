# Function Logic Map: `Repository.apply`

- Source: `internal/protection/repository.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| ID, expected revision, private typed event | ID non-empty, revision positive, event constructed only by event-specific public method | durable stored row + `Transition` | transition/concurrency/storage error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | ID absent | none | `ErrConcurrentUpdate` | invalid input table |
| B2 | revision below one | none | `ErrConcurrentUpdate` | invalid input table |
| B3 | durable load fails | read only | not-found mapped or storage error | missing saga test |
| B4 | load is `sql.ErrNoRows` | none | `ErrConcurrentUpdate` | missing saga test |
| B5 | current revision differs | read only | exact retry or conflict | stale retry/conflict tests |
| B6 | row is exactly prior event result at revision+1 | none | durable row success | same-event idempotency test |
| B7 | pure transition refuses lineage/state/time | none | typed transition error | forged attempt/broker tests |
| B8 | atomic SQL update errors | database attempt only | wrapped storage error | repository error propagation |
| B9 | affected-row inspection errors | none | wrapped storage error | driver contract |
| B10 | CAS affected other than one row | reload only | exact retry or conflict | real two-connection contention |
| B11 | reload is exact revision+1 event result | read only | durable row success | concurrent same-event test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Get`, `Transition`, SQL `WHERE revision=?`, `exactEventResult` | bind caller intent to stored identity/lineage and one atomic revision | no implicit retry; only exact idempotent outcome accepted | CodeGraph + AST |

## State mutations and fallbacks

- Mutates one local saga row after pure transition validation; never calls broker or toggles execution.

## Safety conclusion

- Safe edit boundary: private generic apply is reachable only through event-specific repository methods.
- High-risk impact: yes; durable lineage and concurrency boundary.
