# Function Logic Map: `TestAnIncludedSymbolIsAdoptedWithTheSwitchOff`

- Source: `internal/app/engine/adoption_include_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 47 | `if cycle.Adopted != 1 {` | local/test state assignment | continues through the contract | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff |
| B2 | `if` at line 50 | `if !h.position("005930").Adopted() {` | local/test state assignment | continues through the contract | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff |
| B3 | `if` at line 53 | `if h.position("000660").Adopted() {` | local/test state assignment | continues through the contract | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff |
| B4 | `if` at line 56 | `if cycle.Unmanaged != 1 {` | local/test state assignment | continues through the contract | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDriverHarness` | explicit base-revision dependency at line 40 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `includeOnly` | explicit base-revision dependency at line 41 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `h.holds` | explicit base-revision dependency at line 43 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `h.cycle` | explicit base-revision dependency at line 46 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatalf` | explicit base-revision dependency at line 48 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `Adopted` | explicit base-revision dependency at line 50 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `h.position` | explicit base-revision dependency at line 50 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Error` | explicit base-revision dependency at line 51 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Errorf` | explicit base-revision dependency at line 57 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 3 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
