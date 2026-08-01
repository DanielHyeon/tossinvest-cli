# Branch Test Map: `TestOptimizationRejectsMultipartBeforeReadingPollutedValues`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all multipart fields including invented key are emitted | this test | parser gate absent | PASS |
| B2 | multipart fixture field write succeeds | this test | parser gate absent | PASS |
| B3 | multipart body finalizes | this test | parser gate absent | PASS |
| B4 | authenticated POST request is constructed | this test | parser gate absent | PASS |
| B5 | request reaches console route | this test | parser gate absent | PASS |
| B6 | route returns 400 and calls neither preview nor apply | this test | multipart reached form validation path | PASS |
