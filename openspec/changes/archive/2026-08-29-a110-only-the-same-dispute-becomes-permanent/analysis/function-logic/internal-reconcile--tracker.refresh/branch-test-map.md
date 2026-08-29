# Branch Test Map: `Tracker.Refresh`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | nil authority path | existing Refresh errors | preserve | yes |
| B2 | authority read error retains projection | existing Refresh errors | preserve | yes |
| B3 | durable states are scanned for authority | A110 transient/durable Refresh cases | preserve | yes |
| B4 | foreign durable state is ignored | account isolation | preserve | yes |
| B5 | matching durable permanent is detected | durable Refresh case | preserve | yes |
| B6 | runtime blocks are enumerated | failed-enter A110 tests | preserve | yes |
| B7 | pending runtime block is considered | failed-enter A110 tests | preserve | yes |
| B8 | continuity-broken non-durable account proposal is filtered, ordinary pending remains | `TestA110RefreshDoesNotManufacturePermanentFromContinuityBrokenPendingProposal` | yes | yes |
| B9 | durable rows are rebuilt | account isolation | preserve | yes |
| B10 | foreign durable row is skipped | account isolation | preserve | yes |
| B11 | durable authority or disproved proposal clears pending retry identity | A110 Refresh cases | yes | yes |
| B12 | disproved proposal clears transient streak/failure view | `TestA110RefreshDoesNotManufacturePermanentFromContinuityBrokenPendingProposal` | yes | yes |
| B13 | actual durable permanent lifts compatibility scalar | durable Refresh case | preserve | yes |
