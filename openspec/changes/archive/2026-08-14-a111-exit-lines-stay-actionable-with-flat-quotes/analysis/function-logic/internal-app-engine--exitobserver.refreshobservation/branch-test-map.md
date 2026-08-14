# Branch Test Map: `ExitObserver.refreshObservation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/app/engine/exitloop.go:1050`: recheck monotonic evidence lease at the no-event journal refresh boundary | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `internal/app/engine/exitloop.go:1059`: recheck monotonic evidence lease at the no-event journal refresh boundary | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `internal/app/engine/exitloop.go:1062`: recheck monotonic evidence lease at the no-event journal refresh boundary | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
