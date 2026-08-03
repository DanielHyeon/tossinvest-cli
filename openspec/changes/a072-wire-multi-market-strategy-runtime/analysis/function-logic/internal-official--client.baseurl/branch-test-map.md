# Branch Test Map: `Client.BaseURL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | concurrent option replay cannot race with endpoint observation | TestClientConfigurationOptionReplayIsRaceFreeAndFailClosed | yes (`-race` reported base read/write) | yes |
