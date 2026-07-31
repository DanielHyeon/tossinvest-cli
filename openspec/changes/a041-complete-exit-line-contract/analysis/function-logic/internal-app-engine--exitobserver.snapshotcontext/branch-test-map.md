# Branch Test Map: `ExitObserver.snapshotContext`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | fetched quote context is deterministic and opaque | fetched-at identity test | yes | yes |
| B2 | zero timestamp reuses one cycle fallback | fallback identity test | yes | yes |
