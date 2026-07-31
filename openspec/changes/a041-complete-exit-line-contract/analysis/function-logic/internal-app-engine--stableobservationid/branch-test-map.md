# Branch Test Map: `stableObservationID`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | same fetched quote across different fallbacks | `TestStableObservationIDUsesFetchedAtAndIgnoresCycleFallback` | yes | yes |
| B2 | zero timestamp within/among cycles | `TestStableObservationIDReusesOneFallbackWithinCycle` | yes | yes |
| B3 | raw identity privacy | fetched-at test | yes | yes |
