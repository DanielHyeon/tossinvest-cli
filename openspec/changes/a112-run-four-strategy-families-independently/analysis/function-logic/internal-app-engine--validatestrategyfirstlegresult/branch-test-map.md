# Branch Test Map: `validateStrategyFirstLegResult`

- Source SHA-256: `c31f12fd07855ab32d29c815c8b7b21e14c83add0cd3e1c46503bf01c15eda22`; AST branch locations are authoritative.
- L0 did not alter this function and does not claim an existing test covers a branch.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 114:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B2 | case at 115:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B3 | case at 117:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B4 | case at 119:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B5 | if at 135:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B6 | range at 138:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B7 | if at 139:3 | planned targeted RED before any edit; not run by L0 | no | no |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.
