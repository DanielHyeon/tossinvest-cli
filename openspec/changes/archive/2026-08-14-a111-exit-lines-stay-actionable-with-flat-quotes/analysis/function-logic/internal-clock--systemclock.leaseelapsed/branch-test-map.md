# Branch Test Map: `systemClock.leaseElapsed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free happy path: measure a system lease with time.Since against the retained monotonic reading | `TestSystemLeaseAnchorRetainsMonotonicReading` | preservation or intentional RED as named | focused suite GREEN |
