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
| SQLite query + shared event evidence hydrator | scan every nullable v10 field and validate exact flattened/source/arm evidence | no recomputation; per-event corruption becomes typed unknown | CodeGraph + AST |

## State mutations and fallbacks

- Preserves window ordering and adds persisted evaluation data. Any incomplete or cross-candidate event state clears the effective snapshot view and exposes `invalid_event_evidence` without hiding other account rows.
- Legacy is all-v10-NULL only, and evaluation identity/projection flattening is exact-matched to recomputed JSON.

## Safety conclusion

- Safe edit boundary: read-only console contract; no UI change.
- High-risk impact: yes (ledger read).
