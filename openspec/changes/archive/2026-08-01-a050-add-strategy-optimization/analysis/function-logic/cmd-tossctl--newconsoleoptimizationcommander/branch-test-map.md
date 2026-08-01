# Branch Test Map: `newConsoleOptimizationCommander`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | registry/store construction error is returned without touching the trading journal | `TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore` plus optimization package registry tests | provider was hard-coded nil and could not be wired | yes |
| I1 | complete evidence provider is passed through to the lifecycle | `TestConsoleOptimizationCommanderUsesPerformanceEvidenceProvider` | provider was hard-coded nil | yes |
