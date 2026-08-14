# Branch Test Map: `ExitObserver.judgeRatchet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/app/engine/exitloop.go:917`: recheck the monotonic lease before refresh/full-record dispatch | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `internal/app/engine/exitloop.go:918`: recheck the monotonic lease before refresh/full-record dispatch | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `internal/app/engine/exitloop.go:928`: recheck the monotonic lease before refresh/full-record dispatch | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B4 | `internal/app/engine/exitloop.go:948`: recheck the monotonic lease before refresh/full-record dispatch | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B5 | `internal/app/engine/exitloop.go:956`: recheck the monotonic lease before refresh/full-record dispatch | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B6 | `internal/app/engine/exitloop.go:959`: recheck the monotonic lease before refresh/full-record dispatch | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
