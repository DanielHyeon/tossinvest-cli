# Function Logic Map: `ReadOnly.AccountExitEvents`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account + limit | newest bounded append-only event window | stored event columns | nonpositive limit empty; SQL errors explicit |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | limit guard, query, scan, rows error, success | none | typed result/error | account event read-model tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite query + event scanner | read exact v10 evaluation projection | no recomputation | CodeGraph + AST |

## State mutations and fallbacks

- Preserves window ordering and adds persisted evaluation data.

## Safety conclusion

- Safe edit boundary: read-only console contract; no UI change.
- High-risk impact: yes (ledger read).
