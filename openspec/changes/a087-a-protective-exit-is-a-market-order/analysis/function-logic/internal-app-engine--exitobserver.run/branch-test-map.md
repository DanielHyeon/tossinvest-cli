# Branch Test Map: `ExitObserver.Run`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | completed observation signals refresher before sleep | a087 token-order test | 작성 전 | observe end precedes one refresh token |
| B2 | cancellation before observation | existing Run cancellation test | 기존 green | unchanged ctx error |
| B3 | cancellation during sleep | existing Run cancellation test | 기존 green | unchanged sleep error |
