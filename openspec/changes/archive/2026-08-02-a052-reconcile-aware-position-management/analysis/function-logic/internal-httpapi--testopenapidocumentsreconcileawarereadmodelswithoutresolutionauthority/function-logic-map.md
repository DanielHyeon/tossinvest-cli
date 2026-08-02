# Function Logic Map: `TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority`

- Source: `internal/httpapi/model_openapi_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 97 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B2 | `if` at line 109 | `if err := json.Unmarshal(raw, &document); err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B3 | `range` at line 113 | `for _, name := range []string{"adoptionStatus", "statusKnown", "adoptionLabel", "adoptionReason", "included",` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B4 | `if` at line 115 | `if _, ok := position.Properties[name]; !ok {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B5 | `range` at line 120 | `for _, name := range []string{"desired", "effective", "effectiveKnown", "blockSource"} {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B6 | `if` at line 121 | `if _, ok := management.Properties[name]; !ok {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B7 | `if` at line 125 | `if _, ok := document.Components.Schemas["AdoptionSettings"]; !ok {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B8 | `range` at line 128 | `for path := range document.Paths {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B9 | `if` at line 129 | `if strings.Contains(strings.ToLower(path), "reconcile") {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B10 | `range` at line 133 | `for _, schema := range []string{"ReconcileBlock", "StoredExitEvidence"} {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B11 | `if` at line 135 | `if !ok {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B12 | `range` at line 139 | `for _, forbidden := range []string{"accountRef", "capability", "token", "command"} {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |
| B13 | `if` at line 140 | `if _, leaked := body.Properties[forbidden]; leaked {` | local/read-model state only; see AST assignments | continues through the function contract | TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.ReadFile` | execute the explicit dependency at line 96 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 98 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `json.Unmarshal` | execute the explicit dependency at line 109 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Errorf` | execute the explicit dependency at line 116 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Error` | execute the explicit dependency at line 126 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Contains` | execute the explicit dependency at line 129 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.ToLower` | execute the explicit dependency at line 129 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 9 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
