# Branch Test Map: `systemClock.Since`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free happy path: preserve the frozen-base wall-duration compatibility method while lease helpers own monotonic quote leases | `TestSystemClockSleepShortDuration` | preservation or intentional RED as named | focused suite GREEN |
