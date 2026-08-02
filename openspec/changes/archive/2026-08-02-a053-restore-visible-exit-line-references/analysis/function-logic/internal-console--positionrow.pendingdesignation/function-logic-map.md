# Function Logic Map: `positionRow.PendingDesignation`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Broker presence | present/absent | holdings projection | absent rows are never pending desired adoption |
| Management truth | managed/released/unmanaged/unknown/projected | journal + engine projection | managed, released, unknown, or already-projected rows are false |
| Desired lists | designated/excluded | config snapshot | exclude wins over include |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless conjunction of broker-present, unmanaged, non-released, known, designated, non-excluded, and projection-absent predicates | none | true only for desired-only pending fallback | full truth-table test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Managed` | exclude already-managed rows | pure predicate | portfolio tests |
| `Unknown` | exclude unreadable-journal rows | pure predicate | portfolio notice tests |
| `HasManagementProjection` | defer to typed engine projection | pure predicate | a052/a053 tests |

## State mutations and fallbacks

- Pure boolean projection reused by the management label, absence reason, and exit-reference adapter.
- A known `RELEASED` lifecycle is terminal operator intent for display precedence and cannot be revived by a stale desired include entry.
- Exclude overrides designation and no desired/default percentage or price is read here.

## Safety conclusion

- Safe edit boundary: read-only copy-selection predicate.
- High-risk impact: no direct trading side effect; prevents managed or operator-released rows from being mislabeled as pending.
