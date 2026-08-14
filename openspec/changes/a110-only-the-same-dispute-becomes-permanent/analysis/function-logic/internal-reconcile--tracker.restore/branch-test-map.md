# Branch Test Map: `Tracker.Restore`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | nil journal does not fabricate state | restore error tests | preserve | yes |
| B2 | authority read error does not fabricate state | restore error tests | preserve | yes |
| B3 | active rows are enumerated | restore account isolation | preserve | yes |
| B4 | foreign account rows are skipped | restore account isolation | preserve | yes |
| B5 | durable account quantity row marks permanent | A110 incident/restore tests | yes | yes |
| B6 | permanent compatibility scalar reaches threshold | A110 durable restore test | preserve | yes |
