# Function Logic Map: `Console.handlePositionManagement`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `Console.handlePositionManagement(params=2, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 200 | `if c.opts.Settings == nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B2 | `else` at line 202 | `} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B3 | `if` at line 202 | `} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B4 | `else` at line 204 | `} else {` | local/projection assignment | early return/error nearby | go test ./internal/console |
| B5 | `if` at line 208 | `if c.opts.PositionPolicies == nil {` | local/projection assignment | early return/error nearby | go test ./internal/console |
| B6 | `if` at line 213 | `if runtimeErr != nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B7 | `else` at line 215 | `} else {` | local/projection assignment | continues through contract | go test ./internal/console |
| B8 | `range` at line 218 | `for _, block := range runtime.Blocks {` | local/projection assignment | continues through contract | go test ./internal/console |
| B9 | `if` at line 223 | `if err != nil {` | local/projection assignment | early return/error nearby | go test ./internal/console |
| B10 | `range` at line 228 | `for _, state := range states {` | local/projection assignment | continues through contract | go test ./internal/console |
| B11 | `if` at line 236 | `if management.Block != nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B12 | `if` at line 239 | `if state.Status == positionpolicy.StatusManaged {` | local/projection assignment | continues through contract | go test ./internal/console |
| B13 | `else` at line 248 | `} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecycleEligible() {` | local/projection assignment | continues through contract | go test ./internal/console |
| B14 | `range` at line 241 | `for _, policy := range exitpolicy.RegisteredCommonPolicies() {` | local/projection assignment | continues through contract | go test ./internal/console |
| B15 | `if` at line 245 | `if state.ExternalLifecycleEligible() {` | local/projection assignment | continues through contract | go test ./internal/console |
| B16 | `if` at line 248 | `} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecycleEligible() {` | local/projection assignment | continues through contract | go test ./internal/console |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Get` | explicit dependency at line 198 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.URL.Query` | explicit dependency at line 198 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicy.Descriptor` | explicit dependency at line 198 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.opts.Settings.Load` | explicit dependency at line 202 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `err.Error` | explicit dependency at line 203 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `desiredAdoptionSettingsDisplay` | explicit dependency at line 205 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.render` | explicit dependency at line 209 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.opts.PositionPolicies.Runtime` | explicit dependency at line 212 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.Context` | explicit dependency at line 212 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `runtimeErr.Error` | explicit dependency at line 214 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `adoptionSettingsDisplay` | explicit dependency at line 216 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `append` | explicit dependency at line 219 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newReconcileBlockView` | explicit dependency at line 219 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.now` | explicit dependency at line 219 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.opts.PositionPolicies.List` | explicit dependency at line 222 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicy.ProjectManagement` | explicit dependency at line 229 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.policyAction` | explicit dependency at line 240 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `exitpolicy.RegisteredCommonPolicies` | explicit dependency at line 241 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `state.ExternalLifecycleEligible` | explicit dependency at line 245 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 21 assignment(s), 2 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
