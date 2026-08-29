# Branch Test Map: TestSoakAndLiveEndpointsCoverTheEngineInterlock

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | collect soak/live catalog entries | this test | deployment regression converted to RED | yes |
| B2 | inspect each global endpoint | this test | deployment regression converted to RED | yes |
| B3 | any global endpoint is uncovered | this test | RED failed on exchange-rate GET | yes |
| B4 | inspect global endpoint identities | this test | deployment regression converted to RED | yes |
| B5 | strategy-only exchange GET appears globally | this test | RED failed with endpoint present | yes |
| B6 | WTS evidence falsely covers official OAuth FX | this test | prior contract guard retained | yes |
