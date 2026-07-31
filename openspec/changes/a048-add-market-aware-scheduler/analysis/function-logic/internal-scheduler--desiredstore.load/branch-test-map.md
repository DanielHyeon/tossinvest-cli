# Branch Test Map: `DesiredStore.Load`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | canceled context refuses before read | `TestDesiredStoreHonorsCanceledContext` | context behavior existed | yes |
| success/default | absent file returns revision-0 OFF; valid file round-trips revision | `TestMissingDesiredStateUsesClosedDefaults`, `TestDesiredStateRoundTripsActorApprovalMarketAndVersions` | revision absent | yes |
| delegated refusal | unknown/duplicate/trailing/future approval fails closed | desired load rejection tests | existing | yes |
