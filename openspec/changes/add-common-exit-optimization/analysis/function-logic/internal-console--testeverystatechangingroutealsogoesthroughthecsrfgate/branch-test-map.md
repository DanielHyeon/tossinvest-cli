# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` path at line 385 and its complement/boundary | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
| B2 | `switch` path at line 387 and its complement/boundary | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
| B3 | `case` path at line 388 and its complement/boundary | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
| B4 | `case` path at line 390 and its complement/boundary | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
| B5 | `range` path at line 394 and its complement/boundary | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
| B6 | `if` path at line 395 and its complement/boundary | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
