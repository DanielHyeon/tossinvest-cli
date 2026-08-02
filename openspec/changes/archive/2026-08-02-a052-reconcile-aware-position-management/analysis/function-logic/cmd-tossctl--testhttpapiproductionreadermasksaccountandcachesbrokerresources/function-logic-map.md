# Function Logic Map: `TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources`

- Source: `cmd/tossctl/httpapi_reader_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range` at line 48 | `for range 2 {` | local/test state assignment | continues through the contract | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources |
| B2 | `if` at line 50 | `if err != nil {` | local/test state assignment | continues through the contract | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources |
| B3 | `if` at line 53 | `if len(positions.Items) != 1 \|\| positions.Items[0].Symbol != "005930" \|\|` | local/test state assignment | continues through the contract | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources |
| B4 | `if` at line 58 | `if err != nil {` | local/test state assignment | continues through the contract | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources |
| B5 | `if` at line 61 | `if len(projectedOrders.Items) != 1 \|\| projectedOrders.Items[0].ID != "order-1" \|\|` | local/test state assignment | continues through the contract | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources |
| B6 | `if` at line 66 | `if holdings.calls != 1 \|\| orders.calls != 1 {` | local/test state assignment | continues through the contract | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` | explicit base-revision dependency at line 34 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `reader.Positions` | explicit base-revision dependency at line 49 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `context.Background` | explicit base-revision dependency at line 49 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatal` | explicit base-revision dependency at line 51 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `len` | explicit base-revision dependency at line 53 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `strings.Contains` | explicit base-revision dependency at line 54 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatalf` | explicit base-revision dependency at line 55 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `reader.Orders` | explicit base-revision dependency at line 57 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 6 assignment(s), 2 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
