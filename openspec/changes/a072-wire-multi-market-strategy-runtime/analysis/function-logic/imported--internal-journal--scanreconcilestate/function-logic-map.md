# Function Logic Map: `scanReconcileState`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| database row | reconcileSelect column order including nullable symbol/scope/release values | schema v24 | not-found sentinel or wrapped scan/parse error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no row | none | `ErrReconcileStateNotFound` | inactive release test |
| B2 | scan error | none | wrapped error | schema golden/query tests |
| B3 | invalid entered time | none | wrapped parse error | existing corrupt-row tests |
| B4 | valid optional released time | populate state | state | history/release tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `row.Scan` | materialize durable row | exact select-order contract | AST + schema tests |
| `parseJournalTime` | parse durable timestamps | error fails closed | AST |

## State mutations and fallbacks

- Adds nullable scope-market scan into `ReconcileState`; no database mutation.

## Safety conclusion

- Safe edit boundary: preserve select/scan positional identity while adding one nullable column.
- High-risk impact: yes — read projection feeds reconcile gate/release.
