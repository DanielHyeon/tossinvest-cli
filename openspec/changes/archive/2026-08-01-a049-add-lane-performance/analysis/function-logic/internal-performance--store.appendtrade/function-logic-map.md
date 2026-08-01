# Function Logic Map: `Store.AppendTrade`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| trade | validated immutable derived trade | exact lineage handoff | invalid/divergent bytes fail without write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | validation/begin fails | none | error | store validation tests |
| B2 | identity absent | appends one row in immediate tx | success | collect/store tests |
| B3 | exact identity replay | none | idempotent success | collect replay tests |
| B4 | divergent identity replay | none | `ErrImmutableConflict` | collect divergence tests |
| B5 | commit fails | pending insert rolls back | error | crash/all-or-none suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `compareAndAppendTrade` | exact canonical row equality/insert | no update or fallback | current HEAD + AST |
| `Commit` | atomically publishes one row | error returned | tests |

## State mutations and fallbacks

- Derived trade rows are append-only; no journal, broker, config or LIVE mutation.

## Safety conclusion

- Safe edit boundary: immutable read-model trade append.
- High-risk impact: persistence integrity only.
