# Function Logic Map: `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration`

- Source: `internal/journal/position_policy_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 133 | `if err != nil {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B2 | `if` at line 136 | `if released.Status != positionpolicy.StatusReleased \|\| released.Version != 1 {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B3 | `if` at line 139 | `if released.ObservedAt == "" {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B4 | `if` at line 145 | `if err != nil {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B5 | `if` at line 148 | `if got.AdoptionGeneration != 2 \|\| got.Version != 1 \|\| got.Status != positionpolicy.StatusManaged {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B6 | `if` at line 151 | `if got.EffectivePolicyID != exitpolicy.RatchetPolicyID {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B7 | `if` at line 155 | `if got.ObservedAt == "" \|\| got.ObservedAt == released.ObservedAt {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B8 | `if` at line 162 | `if _, err := j.ApplyPositionPolicy(ctx, late); !errors.Is(err, positionpolicy.ErrVersionMismatch) {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |
| B9 | `if` at line 166 | `if current.AdoptionGeneration != 2 \|\| current.Version != 1 {` | local/test state assignment | continues through the contract | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournal` | explicit base-revision dependency at line 129 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `seedPolicyPosition` | explicit base-revision dependency at line 130 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `context.Background` | explicit base-revision dependency at line 131 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `j.ApplyPositionPolicy` | explicit base-revision dependency at line 132 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `policyRequest` | explicit base-revision dependency at line 132 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatal` | explicit base-revision dependency at line 134 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `t.Fatalf` | explicit base-revision dependency at line 137 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `readoptRequest` | explicit base-revision dependency at line 142 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `At.Add` | explicit base-revision dependency at line 142 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `late.At.Add` | explicit base-revision dependency at line 161 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `errors.Is` | explicit base-revision dependency at line 162 | result/error is handled by the AST-recorded test/function path | base AST + package test |
| `j.PositionPolicy` | explicit base-revision dependency at line 165 | result/error is handled by the AST-recorded test/function path | base AST + package test |

## State mutations and fallbacks

- Base AST records 11 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
