# Branch Test Map: `ExitObserver.judgeLadder`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unknown/replaced policy alerts and does not judge | existing policy-fit tests | existing | yes |
| B2 | no movement produces no event | existing ladder quiet test | existing | yes |
| B3 | one-share intermediate rung records promotion only | a041 exitloop one-share partial test | no | yes |
| B4 | one-share final/breach submit one | a041 exitloop final/breach tests | no | yes |
| B5 | same snapshot has one deterministic proposal under concurrent consumers | a041 race test + existing journal pending invariant | no | yes |
| B6 | changed transition reaches durable record | existing ladder integration tests | existing | yes |
| B7 | snapshot context identity cannot be constructed | decimal/refusal tests | existing | yes |
