# Branch Test Map: `signalsVerdictFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | classify verdict | signals unit suite | existing coverage | one label |
| B2 | raised veto | signals unit suite | existing coverage | 거부 |
| B3 | fully clear | signals unit suite | existing coverage | 통과 |
| B4 | no measurements | `TestARowNobodyMeasuredDoesNotRenderLikeARowThatCleared` | missing accessor during RED | 미확인 |
