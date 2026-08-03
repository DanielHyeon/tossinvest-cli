# Branch Test Map: `TestClientConfigurationOptionReplayIsRaceFreeAndFailClosed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | concurrent replay and origin/base reads | TestClientConfigurationOptionReplayIsRaceFreeAndFailClosed | yes (`-race` reported base/http/origin races) | yes |
| B2 | endpoint remains default after all replay attempts | same test | yes (attacker URL won) | yes |
| B3 | official-origin capability remains valid | same test | yes (origin raced/revoked) | yes |
