# Branch Test Map: `consoleBroker.lock`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | canceled metadata waiter never acquires shared broker state | `TestA061MetadataResolverHonorsCancellationWhileAccountResolutionOwnsTheBroker` | yes | yes |
| tail | concurrent metadata/account reads serialize on one client | `TestA061MetadataAndAccountReadsShareOneOfficialClient` | yes | yes |
