# Function Logic Map: `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint`

- Source: `internal/app/engine/position_policy_transport_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 122 | `if err != nil {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |
| B2 | `if` at line 127 | `if err != nil \|\| info.Mode().Perm()&0o077 != 0 \|\| !info.Mode().IsRegular() {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |
| B3 | `if` at line 131 | `if err != nil {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |
| B4 | `if` at line 135 | `if err != nil \|\| len(states) != 1 \|\| states[0].Version != 4 {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |
| B5 | `if` at line 138 | `if err := server.Close(); err != nil {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |
| B6 | `if` at line 141 | `if _, err := os.Stat(descriptorPath); !errors.Is(err, os.ErrNotExist) {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |
| B7 | `if` at line 144 | `if entries, err := os.ReadDir(dir); err != nil \|\| len(entries) != 0 {` | local/test state assignment | continues through the contract | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `privateEngineTestDir` | explicit base-revision dependency at line 116 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `StartPositionPolicyCommandServer` | explicit base-revision dependency at line 121 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatal` | explicit base-revision dependency at line 123 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `positionpolicyrpc.DescriptorPath` | explicit base-revision dependency at line 125 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `os.Lstat` | explicit base-revision dependency at line 126 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `Perm` | explicit base-revision dependency at line 127 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `info.Mode` | explicit base-revision dependency at line 127 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `IsRegular` | explicit base-revision dependency at line 127 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatalf` | explicit base-revision dependency at line 128 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `positionpolicyrpc.Dial` | explicit base-revision dependency at line 130 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `context.Background` | explicit base-revision dependency at line 130 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `client.List` | explicit base-revision dependency at line 134 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `len` | explicit base-revision dependency at line 135 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `server.Close` | explicit base-revision dependency at line 138 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `os.Stat` | explicit base-revision dependency at line 141 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `errors.Is` | explicit base-revision dependency at line 141 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `os.ReadDir` | explicit base-revision dependency at line 144 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 10 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
