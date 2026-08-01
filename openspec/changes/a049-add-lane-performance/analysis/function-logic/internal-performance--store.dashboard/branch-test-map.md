# Branch Test Map: `Store.Dashboard`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | period zero defaults; >90 refuses | dashboard bounds test | yes | yes |
| B2 | empty filters become fixed defaults | default query test | existing | existing |
| B3 | >10k rows refuses without unbounded allocation | row-bound test | yes | yes |
| B4 | bounded rows aggregate metrics/provenance | dashboard tests | existing | existing |
| B5 | StateCounts use identical market/lane/complete predicates | filter parity test | yes | yes |
| B6 | only filtered trades contribute missing state | filter parity/missing tests | yes | yes |
| B7 | all six metrics are checked | missing-state test | no — existing | yes |
| B8 | one missing metric counts trade once | missing-state test | no — existing | yes |
