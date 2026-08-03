# Branch Test Map: `NewRouter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil primary reader refused | existing router constructor contract | pending | pending |
| B2 | nil clock defaults | existing stable envelope test | pending | pending |
| B3 | unsafe mutation path refused | `TestRouterConfigurationRejectsUnsafeMutationPaths` | pending | pending |
| B4 | optional strategy reader retained without mutation authority | `TestStrategyRuntimeRESTUsesSharedProjectionAndStrictGuards` | pending | pending |
