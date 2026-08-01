# Branch Test Map: `scanExitEvent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | positional scan error propagates | `exit snapshot persistence suite` | yes | yes |
| B2 | zero/negative lifecycle generation is corrupt | `TestExitEventsCarryLifecycleGeneration` | yes | yes |
