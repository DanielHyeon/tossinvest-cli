# Branch Test Map: `validateStrategyFirstLegResult`

- Source SHA-256: `3c82793300b39454ccac2ff41fe97c0b76ed2190534c5a76c0f9f7abd59652e5`; AST branch locations are authoritative.
- L0 did not alter this function and does not claim an existing test covers a branch.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 111:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B2 | case at 112:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B3 | case at 114:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B4 | case at 116:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B5 | if at 132:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B6 | range at 135:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B7 | if at 136:3 | planned targeted RED before any edit; not run by L0 | no | no |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.
