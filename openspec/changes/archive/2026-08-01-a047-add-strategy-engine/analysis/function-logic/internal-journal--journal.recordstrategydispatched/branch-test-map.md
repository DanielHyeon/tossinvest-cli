# Branch Test Map: `Journal.RecordStrategyDispatched`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | PLANNED confirmed outcome | confirmed recovery row | missing path | pass |
| B2 | IN_DOUBT later confirmed | `TestStrategyRecoveryCanPromoteCurrentInDoubtReceiptToDispatched` | PLANNED-only CAS | pass |
| B3 | stale/forged receipt | terminal CAS tests | partial binding | pass |
| B4 | exact execution links reverse-resolve | restart reverse lookup test | missing links | pass |
| B5 | transaction/link insert error rolls back state | collision/trigger tests | missing | pass |
| B6 | reverse external ref points to same attempt | reverse lookup tests | weak draft | pass |
| B7 | reverse ref collision refuses | divergent execution test | weak draft | pass |
| B8 | exact link pair commits | confirmed recovery test | missing | pass |
| B9 | exact receipt binding precheck fails | stale/forged terminal tests | immutable identity named in UPDATE | pass |
| B10 | already DISPATCHED exact retry converges | dispatch idempotency/recovery tests | stale retry risk | pass |
