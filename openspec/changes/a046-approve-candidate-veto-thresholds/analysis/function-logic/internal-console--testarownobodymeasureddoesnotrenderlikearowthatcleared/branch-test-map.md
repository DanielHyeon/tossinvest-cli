# Branch Test Map: `TestARowNobodyMeasuredDoesNotRenderLikeARowThatCleared`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | wrong unchecked label | this test | existing guard | 미확인 |
| B2 | wrong measured detail | this test | existing guard | 0/3 |
| B3 | wrong cell count | this test | existing guard | three |
| B4 | inspect missing cells | this test | existing guard | all inspected |
| B5 | missing appears clear | this test | existing guard | reason only |
| B6 | inspect clear cells | this test | existing guard | all inspected |
| B7 | clear cell malformed | this test | existing guard | clear only |
| B8 | wrong passed label | this test | existing guard | 통과 |
| B9 | clear count drift | this test | missing accessor during RED | equals copied order length |
