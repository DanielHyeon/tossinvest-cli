# Function Logic Map: `Journal.ReleaseReconcile`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| release request | normalized account, optional symbol, valid release cause/evidence | request + release contract | `ErrInvalidRequest`, no write |
| scope market | empty global; KR/US exact symbol scope only | v24 scope contract | invalid refused; exact lookup never falls across market |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid request/scope | none | `ErrInvalidRequest` | validation tests |
| B2 | exact scope absent | read only | `released=false` | cross-market refusal test |
| B3 | expected cause differs | read only | active row, `released=false` | existing cause-ownership test |
| B4 | exact row matches | UPDATE + commit | released row | exact-market release test |
| B5 | storage failure | rollback | wrapped error | existing journal error coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanReconcileState` | read exact release target | fail closed on storage/parse error | CodeGraph + AST |
| SQLite transaction | atomically select/update | rollback on error; no retry | CodeGraph + AST |

## State mutations and fallbacks

- Updates only the selected row ID. A market-specific request cannot select legacy global or the peer market.

## Safety conclusion

- Safe edit boundary: add validated market to selection while preserving exact ID update and cause ownership.
- High-risk impact: yes — release reduces a durable trading block.
