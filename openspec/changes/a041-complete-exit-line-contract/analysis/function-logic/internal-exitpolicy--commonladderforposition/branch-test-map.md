# Branch Test Map: `CommonLadderForPosition`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unknown id fails | existing common registry test | existing | yes |
| B2 | adopted RUNNER has all partials zero and distinct valid identity | existing adopted runner + a041 identity tests | existing | yes |
| B3 | self-entered RUNNER and other policies preserve exact tables | existing common policy tests | existing | yes |
| B4 | rewritten adopted RUNNER receives a distinct valid identity | existing adopted runner + a041 identity tests | no | yes |
| B5 | adopted RUNNER pinned digest drift is refused | pinned semantic digest test | yes | yes |
