# Branch Test Map: `applyStoredPositionForRequest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `cmd/tossctl/httpapi_reader.go:284`: preserve stable machine stale reasons while consuming shared response authority | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `cmd/tossctl/httpapi_reader.go:287`: preserve stable machine stale reasons while consuming shared response authority | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
