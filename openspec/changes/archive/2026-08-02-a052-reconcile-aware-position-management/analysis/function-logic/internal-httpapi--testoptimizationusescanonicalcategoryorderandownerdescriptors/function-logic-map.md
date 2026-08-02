# Function Logic Map: `TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors`

- Source: `internal/httpapi/router_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 114 | `if err != nil {` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B2 | `range` at line 119 | `for _, category := range resource.Categories {` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B3 | `if` at line 123 | `if !reflect.DeepEqual(gotIDs, wantIDs) {` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B4 | `if` at line 126 | `if resource.PositionManagement.AutoEnabledDefault \|\| resource.PositionManagement.StopDefault != "5%" \|\|` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B5 | `if` at line 130 | `if len(resource.CandidateFilters) != 2 \|\| len(resource.CandidateFilters[0].Filters) == 0 {` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B6 | `if` at line 134 | `if first.DefaultState != "unapproved" \|\| first.DesiredValue != "" \|\| first.EffectiveValue != "" {` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B7 | `if` at line 137 | `if len(resource.Fields) != 1 \|\| resource.Fields[0].Key != "exit.common-policy" \|\| resource.Fields[0].Owner != "a041-complete-exit-line-contract" {` | local/test state assignment | continues through the contract | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Optimization` | explicit base-revision dependency at line 113 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `context.Background` | explicit base-revision dependency at line 113 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatal` | explicit base-revision dependency at line 115 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `OptimizationFrom` | explicit base-revision dependency at line 117 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `make` | explicit base-revision dependency at line 118 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `len` | explicit base-revision dependency at line 118 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `append` | explicit base-revision dependency at line 120 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `reflect.DeepEqual` | explicit base-revision dependency at line 123 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatalf` | explicit base-revision dependency at line 124 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 6 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
