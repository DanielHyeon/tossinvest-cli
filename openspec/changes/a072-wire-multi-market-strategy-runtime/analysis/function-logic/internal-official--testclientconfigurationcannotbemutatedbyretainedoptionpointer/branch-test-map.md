# Branch Test Map: `TestClientConfigurationCannotBeMutatedByRetainedOptionPointer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | option receives exact construction pointer | TestClientConfigurationCannotBeMutatedByRetainedOptionPointer | yes | yes |
| B2 | all post-construction option writes are ignored | same test | yes (base/http/account changed) | yes |
| B3 | sealed production origin remains authoritative | same test | yes (replay revoked origin) | yes |
