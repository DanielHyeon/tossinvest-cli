# Branch Test Map: `Repository.apply`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty ID | repository invalid-input test | arbitrary update accepted IDs from saga | rejected |
| B2 | nonpositive expected revision | repository invalid-input test | full saga revision drove update | rejected |
| B3 | load error | missing/corrupt row test | no stored-event API | propagated |
| B4 | missing row | missing saga test | no stored-event API | concurrency error |
| B5 | stale revision with different result | conflicting update test | caller state could race | conflict |
| B6 | stale revision with exact result | round-trip retry test | no event identity retry | durable result returned |
| B7 | invalid state/attempt/broker lineage | forged lineage tests | mismatches accepted | typed refusal |
| B8 | SQL execution failure | storage error test | n/a | propagated |
| B9 | row-count failure | driver contract test | n/a | propagated |
| B10 | real concurrent CAS loss | two independent SQLite connections | one-connection test serialized | exactly one conflicting event wins |
| B11 | same event wins before peer | concurrent same-event test | no exact retry recognition | both callers succeed, one revision |
