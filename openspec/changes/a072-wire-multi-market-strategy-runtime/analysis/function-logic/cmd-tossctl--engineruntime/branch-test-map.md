# Branch Test Map: `engineRuntime`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | hardcoded nil-hint fill detector path and detector failure | existing engine runtime branch tests | existing | GREEN |
| B2 | reconcile construction failure | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | existing | GREEN |
| B3 | exit observer construction failure | same | existing | GREEN |
| B4 | recovery construction failure | same | existing | GREEN |
| B5 | paired dormant supervisor construction cannot receive configuration and failure aborts assembly | dormant production assembly tests | RED before wiring | GREEN |
| S1 | successful runtime loop names are reconcile, exit, filldetect, strategy-entry | `TestProductionRuntimeIncludesOneDormantStrategyEntryOuterLoop` | RED before wiring | GREEN |
| S2 | dormant KR/US are OFF, Trigger disabled, no Cycle and cancellation drains | `TestProductionStrategyEntryAssemblyIsPairedDormantAndInert` | RED before wiring | GREEN |
| S3 | source/import boundary contains no activation, Gateway, journal mutation, toggle or LIVE input in dormant helper | `TestDormantProductionHelperHasNoAuthorityOrMutationInput` | RED before wiring | GREEN |
