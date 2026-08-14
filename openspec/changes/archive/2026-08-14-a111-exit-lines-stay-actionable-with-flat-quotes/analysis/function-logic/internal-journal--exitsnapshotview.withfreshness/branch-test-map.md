# Branch Test Map: `ExitSnapshotView.WithFreshness`

| Branch | Scenario | Test/evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil snapshot, disabled age, or zero as-of | freshness no-op table | existing | existing |
| B2 | observed-at parse fails | invalid time case | existing | existing |
| B3 | observation older than limit | old snapshot case | existing | existing |
| B4 | first age comparison false, evaluate future branch | within-limit/future table | existing | existing |
| B5 | observation is after as-of | future snapshot case | existing | existing |

A111 writer integration: refreshed persisted `ObservedAt` remains fresh inside and exactly at 30 seconds, and becomes stale only after valid polling stops beyond the bound.
