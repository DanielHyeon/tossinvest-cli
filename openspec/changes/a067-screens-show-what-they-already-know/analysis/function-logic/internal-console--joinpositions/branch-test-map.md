# Branch Test Map: `joinPositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | every broker holding becomes a row | existing positions tests | n/a | pass |
| B2 | ledger rows merge onto the matching holding; the quarantine travels with the row it belongs to | existing positions tests; `TestAQuarantinedPositionIsNotDrawnAsProtected` | yes | yes |
| B3 | a ledger position the broker does not report keeps its own row | existing broker-missing test | n/a | pass |
| B4 | a second live instance of one symbol is not overwritten | existing duplicate-instance test | n/a | pass |
| B5 | same condition, other arm | same | n/a | pass |
| B6 | rows sort by market then symbol | existing ordering tests | n/a | pass |
