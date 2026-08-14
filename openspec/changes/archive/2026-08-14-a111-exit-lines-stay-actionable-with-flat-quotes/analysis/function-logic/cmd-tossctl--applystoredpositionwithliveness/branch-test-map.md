# Branch Test Map: `applyStoredPositionWithLiveness`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `cmd/tossctl/httpapi_reader.go:298`: expose actionable fields only from canonical persisted evidence accepted by shared freshness | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `cmd/tossctl/httpapi_reader.go:300`: expose actionable fields only from canonical persisted evidence accepted by shared freshness | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `cmd/tossctl/httpapi_reader.go:305`: expose actionable fields only from canonical persisted evidence accepted by shared freshness | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B4 | `cmd/tossctl/httpapi_reader.go:308`: expose actionable fields only from canonical persisted evidence accepted by shared freshness | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B5 | `cmd/tossctl/httpapi_reader.go:312`: expose actionable fields only from canonical persisted evidence accepted by shared freshness | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B6 | `cmd/tossctl/httpapi_reader.go:312`: expose actionable fields only from canonical persisted evidence accepted by shared freshness | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
