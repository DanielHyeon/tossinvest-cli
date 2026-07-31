# Branch Test Map: `DecideFlatten`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | required in follow-up RED | pending implementation |
| B1+ | Existing code only checks deadline and quantity; it omits start/order/duration/scope/identity. | valid boundary, >2s, reversed timestamps, mismatched scope/broker, trigger race, insufficient sellable. | required in follow-up RED | pending implementation |
