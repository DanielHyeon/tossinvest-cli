# Branch Test Map: `ExitObserver.judgeRatchet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid policy/input alerts and performs no write/order | existing refusal tests | existing | yes |
| B2 | unchanged snapshot produces no event | existing quiet-cycle tests | existing | yes |
| B3 | one-share partial records promotion with zero journal proposal/broker call | a041 exitloop one-share test | no | yes |
| B4 | one-share breach records and submits quantity one | a041 exitloop breach test | no | yes |
| B5 | deterministic same input yields same snapshot/decision ID | a041 deterministic ID test | no | yes |
