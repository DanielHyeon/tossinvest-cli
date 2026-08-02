# Function Logic Map: `TestEveryOpenAPILocalReferenceResolves`

- Source: `internal/httpapi/model_openapi_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestEveryOpenAPILocalReferenceResolves(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 66 | `if err != nil {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B2 | `if` at line 70 | `if err := json.Unmarshal(raw, &document); err != nil {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B3 | `type-switch` at line 75 | `switch typed := value.(type) {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B4 | `case` at line 76 | `case map[string]any:` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B5 | `range` at line 77 | `for key, child := range typed {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B6 | `if` at line 78 | `if key == "$ref" {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B7 | `if` at line 80 | `if !ok \|\| !localReferenceExists(document, ref) {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B8 | `case` at line 86 | `case []any:` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |
| B9 | `range` at line 87 | `for _, child := range typed {` | local/test state assignment | continues through the contract | TestEveryOpenAPILocalReferenceResolves |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.ReadFile` | explicit base-revision dependency at line 65 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatal` | explicit base-revision dependency at line 67 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `json.Unmarshal` | explicit base-revision dependency at line 70 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `localReferenceExists` | explicit base-revision dependency at line 80 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Errorf` | explicit base-revision dependency at line 81 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `walk` | explicit base-revision dependency at line 84 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 5 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
