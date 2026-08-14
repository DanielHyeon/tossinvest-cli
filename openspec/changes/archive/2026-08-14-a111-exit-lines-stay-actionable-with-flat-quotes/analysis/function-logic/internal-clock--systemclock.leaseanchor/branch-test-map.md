# Branch Test Map: `systemClock.leaseAnchor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free happy path: retain Go time.Time monotonic reading for a process-local nonpersisted anchor | `TestSystemLeaseAnchorRetainsMonotonicReading` | preservation or intentional RED as named | focused suite GREEN |
