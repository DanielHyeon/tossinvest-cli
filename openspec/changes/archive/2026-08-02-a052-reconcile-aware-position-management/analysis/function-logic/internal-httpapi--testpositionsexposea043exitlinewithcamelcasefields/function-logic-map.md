# Function Logic Map: `TestPositionsExposeA043ExitLineWithCamelCaseFields`

- Source: `internal/httpapi/router_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestPositionsExposeA043ExitLineWithCamelCaseFields(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range` at line 100 | `for _, want := range []string{'"exitLine"', '"currentProtection":"68000"', '"statusText":"평가 완료"'} {` | local/test state assignment | continues through the contract | TestPositionsExposeA043ExitLineWithCamelCaseFields |
| B2 | `if` at line 101 | `if !strings.Contains(body, want) {` | local/test state assignment | continues through the contract | TestPositionsExposeA043ExitLineWithCamelCaseFields |
| B3 | `range` at line 105 | `for _, forbidden := range []string{'"ExitLine"', '"CurrentProtection"', '"StatusText"'} {` | local/test state assignment | continues through the contract | TestPositionsExposeA043ExitLineWithCamelCaseFields |
| B4 | `if` at line 106 | `if strings.Contains(body, forbidden) {` | local/test state assignment | continues through the contract | TestPositionsExposeA043ExitLineWithCamelCaseFields |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewRouter` | explicit base-revision dependency at line 96 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `httptest.NewRecorder` | explicit base-revision dependency at line 97 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `handler.ServeHTTP` | explicit base-revision dependency at line 98 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `httptest.NewRequest` | explicit base-revision dependency at line 98 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `recorder.Body.String` | explicit base-revision dependency at line 99 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `strings.Contains` | explicit base-revision dependency at line 101 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Errorf` | explicit base-revision dependency at line 102 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 3 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
