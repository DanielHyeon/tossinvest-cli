# Function Logic Map: `ReadOnly.LivePositionExits`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account | non-closed positions with optional typed exit/snapshot projection | persisted journal | unknown is rendered as reason, never default/recompute |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | state query, position query, merge, row error, unknown/full projection, success | none | typed read model/error | account snapshot tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| account snapshot result reader | merge by exact position id | no fuzzy join/recompute | CodeGraph + AST |

## State mutations and fallbacks

- Semantic corruption is kept as a typed unknown view for the console; storage failures remain errors.

## Safety conclusion

- Safe edit boundary: read-only projection.
- High-risk impact: yes (ledger read).
