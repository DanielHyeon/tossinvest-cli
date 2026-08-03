# Branch Test Map: `WithHTTPClient`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | post-construction HTTP option replay is locked and ignored | `TestClientConfigurationOptionReplayIsRaceFreeAndFailClosed` | yes (race detector reported transport/origin races) | yes |
