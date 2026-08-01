# Branch Test Map: `TestAnAbsentThresholdIsNotAPassedVeto`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | absent thresholds pass | this test | existing guard | rejected |
| B2 | inspect all codes | this test | missing accessor during RED | all three |
| B3 | absent code appears measured | this test | existing guard | unmeasured |
| B4 | wrong reason | this test | existing guard | THRESHOLD_ABSENT |
