# Branch Test Map: `TestEveryAssessedCandidateLandsInExactlyOneBucket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | total count | this test | existing guard | expected total |
| B2 | passed count | this test | existing guard | expected passed |
| B3 | vetoed count | this test | existing guard | expected vetoed |
| B4 | unmeasured count | this test | existing guard | expected unmeasured |
| B5 | exclusive sum | this test | existing guard | equals total |
| B6 | near-high missing | this test | existing guard | expected count |
| B7 | seen-late raised | this test | existing guard | expected count |
| B8 | all copied D3 codes | this test | missing accessor during RED | three keys |
| B9 | Raised map key | this test | existing guard | present |
| B10 | NotMeasured map key | this test | existing guard | present |
| B11 | reason tally | this test | existing guard | expected count |
