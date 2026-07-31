# Function Logic Map: `insertAttemptWithBrokerOrder`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test journal and ids | deterministic fixture | journal tests | fatal on insert failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | insert failure | test-only DB write | fatal test | journal focused tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite insert | create attempt fixture | test fails immediately | AST + package tests |

## State mutations and fallbacks

- Test helper only; production code unreachable.

## Safety conclusion

- Safe edit boundary: test fixture.
- High-risk impact: no.
