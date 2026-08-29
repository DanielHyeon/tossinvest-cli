# Function Logic Map: `Journal.ReleaseReconciles`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request batch | non-empty valid release requests, each with expected cause | release contract | all invalid/missing/ownership cases fail before mutation |
| scope identity | account + symbol + market; empty market is global | v24 scope contract | duplicate exact scope refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | empty batch | none | nil | existing empty-batch behavior |
| B2 | invalid/duplicate normalized scope | none | `ErrInvalidRequest` | validation/duplicate tests |
| B3 | any exact scope missing or wrong cause | read only, rollback | error | atomic rollback and cross-market tests |
| B4 | all scopes valid | UPDATE each + commit | released states | atomic release tests |
| B5 | update row count/storage/commit error | rollback | wrapped error | existing storage contract |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanReconcileState` | prove every exact target before update | any failure aborts batch | CodeGraph + AST |
| SQLite transaction | all-or-nothing releases | rollback on any error; no retry | CodeGraph + AST |

## State mutations and fallbacks

- Duplicate key includes market. Selection is exact; legacy global and KR/US are distinct release scopes.

## Safety conclusion

- Safe edit boundary: extend normalization/dedup/select only; retain preflight-before-mutation and row-count checks.
- High-risk impact: yes — atomic release of durable trading guards.
