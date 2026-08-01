# Branch Test Map: `appendSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | new identity appends header+six metrics | collect test | existing | existing |
| B2 | exact snapshot replay skips | replay/restart/concurrent tests | yes | yes |
| B3 | same identity divergent metric fails | divergence test | yes | yes |
| B4 | SIGKILL during snapshot append leaks no partial rows | phase crash test | yes | yes |
| B5 | new header insert failure aborts | crash/all-or-none suite | yes | yes |
| B6 | identity retrieval failure aborts | driver/error contract | no — defensive | yes |
| B7 | exactly six canonical metrics append | collect row-count test | no — existing | yes |
| B8 | any metric failure rolls back header and prior metrics | phase crash test | yes | yes |
