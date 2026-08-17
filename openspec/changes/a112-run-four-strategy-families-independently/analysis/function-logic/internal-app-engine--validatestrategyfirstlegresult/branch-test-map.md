# Branch Test Map: `validateStrategyFirstLegResult`

- Source SHA-256: `a08618229629b30fd7f4f45b19b3773cb9b1e84f9dc3eebf6654e44ea4e72894`; AST branch locations are authoritative.
- L0 did not alter this function and does not claim an existing test covers a branch.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 106:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B2 | case at 107:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B3 | case at 109:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B4 | case at 111:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B5 | if at 127:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B6 | range at 130:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B7 | if at 131:3 | planned targeted RED before any edit; not run by L0 | no | no |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.
