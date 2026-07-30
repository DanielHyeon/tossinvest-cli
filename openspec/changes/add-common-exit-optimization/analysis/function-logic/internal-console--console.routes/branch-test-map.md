# Branch Test Map: `Console.routes`

- Source: `internal/console/console.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy path | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests | yes | yes |
