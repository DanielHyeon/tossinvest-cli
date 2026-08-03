# Branch Test Map: `WithAccountSeq`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | retained pointer cannot replay account selection after seal | TestClientConfigurationCannotBeMutatedByRetainedOptionPointer | yes (account changed to 99) | yes |
