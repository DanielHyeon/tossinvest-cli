# Branch Test Map: `engineArgs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/empty root | existing engineArgs default test | yes | yes |
| B2 | explicit config dir | existing profile propagation test | yes | yes |
| B3 | explicit session file | new `TestEngineArgsCarryExplicitSessionFile` | no | no |
| B4 | both explicit | new test verifies stable global flag order | no | no |
