# Branch Test Map: `WithBaseURL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | post-construction base option replay is locked and ignored | `TestClientConfigurationCannotBeMutatedByRetainedOptionPointer` | yes (base changed to attacker origin) | yes |
