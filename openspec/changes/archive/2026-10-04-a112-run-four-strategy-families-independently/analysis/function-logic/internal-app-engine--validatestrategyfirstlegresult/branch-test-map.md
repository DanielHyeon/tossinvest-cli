# Branch Test Map: `validateStrategyFirstLegResult`

- Source SHA-256: `254b4a6abb0d95febd036b0f437829c391e61000fa71b9b9abac1241ee14444c`; AST branch locations are authoritative.
- L0 did not alter this function and does not claim an existing test covers a branch.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 123:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B2 | case at 124:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B3 | case at 126:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B4 | case at 128:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B5 | if at 144:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B6 | range at 147:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B7 | if at 148:3 | planned targeted RED before any edit; not run by L0 | no | no |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.
