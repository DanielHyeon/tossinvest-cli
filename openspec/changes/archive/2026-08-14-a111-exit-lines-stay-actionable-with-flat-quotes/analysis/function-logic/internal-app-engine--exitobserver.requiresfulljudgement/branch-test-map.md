# Branch Test Map: `ExitObserver.requiresFullJudgement`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/app/engine/exitloop.go:1041`: route SEED, rejudge, missing effective or semantic movement through the full judgement transaction | `TestA111PendingSuppressionSemanticChangeUsesFullJudgementPath` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
