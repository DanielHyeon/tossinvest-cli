# Branch Test Map: `buildGateway`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | existing projection/restore failure | engine startup regression tests | baseline | yes |
| B2 | no fill lifecycle keeps KR/US UNWIRED | `TestUnprovenFillLifecycleKeepsBothProductionAssembliesUnwired` | claimed WIRED | yes |
| B3 | arbitrary official factory/DB absent | `TestEngineHasNoArbitraryProtectionGatewayFactoryOrProtectionDBStartupDependency` | present | yes |
| B4 | protection DB collision does not stop safety runtime | `TestProtectionStorageFailureCannotPreventSafetyRuntimeAssembly` | startup dependency present | yes |
