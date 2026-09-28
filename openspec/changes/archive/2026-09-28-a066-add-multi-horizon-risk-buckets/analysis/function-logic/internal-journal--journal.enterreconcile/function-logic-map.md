# Function Logic Map: `Journal.EnterReconcile`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account/symbol/cause/evidence | normalized account, optional symbol, enumerated cause, non-empty evidence | request + journal contract | `ErrInvalidRequest`, no write |
| scope market | empty means legacy/global; otherwise KR or US and requires symbol | v24 reconcile scope contract | `ErrInvalidRequest`, no write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid account/cause/evidence/market combination | none | `ErrInvalidRequest` | validation tests |
| B2 | global or same-market active row exists | read only | existing row, `entered=false` | global-blocks-market and re-entry tests |
| B3 | no active overlapping row | INSERT + commit | new row, `entered=true` | independent KR/US entry test |
| B4 | begin/read/insert/commit fails | rollback | wrapped storage error | migration/index tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanReconcileState` | read existing overlap | fail closed on scan/storage error | CodeGraph + AST |
| SQLite transaction | serialize active-scope check and insert | rollback on any error; no retry | CodeGraph + AST |

## State mutations and fallbacks

- Writes one active global or exact-market row. Global NULL overlaps every market; exact KR and US may coexist. No live broker side effect.

## Safety conclusion

- Safe edit boundary: carry validated scope through lookup, ID, INSERT and return without changing cause/evidence semantics.
- High-risk impact: yes — durable reconciliation admission guard.
