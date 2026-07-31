# Branch Test Map: `TallyVetoes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero tally still has three code keys | `TestEveryAssessedCandidateLandsInExactlyOneBucket` | existing coverage | three keys |
| B2 | multiple candidates | same test | existing coverage | all consumed |
| B3 | each candidate visits three codes | same test | exported array mutable | copy accessor preserves three |
| B4 | state classification | veto tests | existing coverage | exclusive classification |
| B5 | raised state | raised tally assertions | existing coverage | Raised incremented |
| B6 | unmeasured state | missing tally assertions | existing coverage | missing/reason incremented |
| B7 | candidate partition | bucket test | existing coverage | exactly one bucket |
| B8 | candidate vetoed | bucket test | existing coverage | Vetoed incremented |
| B9 | candidate unmeasured | absent threshold test | existing coverage | Unmeasured incremented |
| B10 | candidate clear | passed test | existing coverage | Passed incremented |
