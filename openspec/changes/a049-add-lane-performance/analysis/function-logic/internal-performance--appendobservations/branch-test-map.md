# Branch Test Map: `appendObservations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid observation rejected | model/store validation tests | existing | existing |
| B2 | new immutable row inserted | compare-and-append test | yes | yes |
| B3 | identical row skipped | replay/restart/concurrent tests | yes | yes |
| B4 | divergent row refused | divergence tests | yes | yes |
| B5 | late overdue append clears completed cadence and prunes immediately | `TestPruneKeepsCadenceDueUntilBoundedBacklogDrainsDespiteContinuedInflux` | yes | yes |
| B6 | INSERT error aborts batch | crash/all-or-none tests | yes | yes |
| B7 | cadence metadata mutation is transactional | continuous influx + crash tests | yes | yes |
