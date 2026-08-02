# Function Logic Map: `Console.overview`

- Source: `internal/console/overview.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Request context and clock | one authenticated dashboard request | handler context and injected clock | cancellation/failure belongs to individual panel state |
| Ledger read | live positions, trips, events and one shared verdict | `overviewLedger` single read-only handle | partial facts retained and failure named |
| Panels | engine, account, today, recent, orders, safety, verify | independent builders | one panel failure does not blank the others |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Branchless assembly in fixed order | fills request-local `overviewView` | one return; failures represented in panels | overview independence and shared holdings tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `overviewLedger` | one consistent journal snapshot | one read-only open, no retry | AST + ledger-open-once test |
| `accountPanelFrom` | zero-call cached holdings panel | must keep `peek` behavior | CodeGraph impact + rate-budget tests |
| remaining panel builders | preserve existing overview facts | independent typed missing states | overview test suite |

## State mutations and fallbacks

- The function only assembles a request-local view.
- Passing request context into account row enrichment must not cause a broker refresh or mutation.

## Safety conclusion

- Safe edit boundary: thread context into the account builder while preserving single-ledger-read and independent panels.
- High-risk impact: no; read-only display assembly only.
