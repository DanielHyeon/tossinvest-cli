# Branch Test Map: `Journal.ExitEvents`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query failure is wrapped | `exit event suite` | yes | yes |
| B2 | scan every ordered event | `TestExitEventsCarryLifecycleGeneration` | yes | yes |
| B3 | scan corruption fails closed | `exit snapshot corruption suite` | yes | yes |
| B4 | iterator failure propagates | `exit event suite` | yes | yes |
