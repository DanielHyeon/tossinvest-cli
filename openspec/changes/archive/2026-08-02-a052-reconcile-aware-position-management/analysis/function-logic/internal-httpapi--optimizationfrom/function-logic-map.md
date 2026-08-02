# Function Logic Map: `OptimizationFrom`

- Source: `internal/httpapi/read.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `OptimizationFrom(params=1, results=1)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 239 | `if out.Evidence.Missing == nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B2 | `if` at line 242 | `if !view.Evidence.ObservedAt.IsZero() {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B3 | `range` at line 246 | `for _, category := range optimization.Categories() {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B4 | `range` at line 250 | `for _, registered := range view.Registry.All() {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B5 | `range` at line 260 | `for _, option := range descriptor.Options {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B6 | `range` at line 266 | `for _, snapshot := range view.History {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |
| B7 | `range` at line 269 | `for _, event := range view.Audit {` | local/read-model state only; see AST assignments | early return/error path nearby | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `make` | execute the explicit dependency at line 229 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `len` | execute the explicit dependency at line 229 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `optimization.Categories` | execute the explicit dependency at line 229 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `view.Registry.All` | execute the explicit dependency at line 230 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `positionManagementFrom` | execute the explicit dependency at line 231 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `positionpolicy.Descriptor` | execute the explicit dependency at line 231 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `candidateFiltersFrom` | execute the explicit dependency at line 232 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `candidate.CandidateFilterMarkets` | execute the explicit dependency at line 232 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `runtimeDescriptorFrom` | execute the explicit dependency at line 233 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strategyengine.DormantRuntimeDescriptor` | execute the explicit dependency at line 233 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `append` | execute the explicit dependency at line 235 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `call` | execute the explicit dependency at line 235 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `view.Evidence.ObservedAt.IsZero` | execute the explicit dependency at line 242 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `view.Evidence.ObservedAt.UTC` | execute the explicit dependency at line 243 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `string` | execute the explicit dependency at line 247 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `stateFrom` | execute the explicit dependency at line 254 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `stateForOption` | execute the explicit dependency at line 255 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `provenanceFrom` | execute the explicit dependency at line 259 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `optimizationSnapshotFrom` | execute the explicit dependency at line 267 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `event.CreatedAt.UTC` | execute the explicit dependency at line 273 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 12 assignment(s), 1 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
