# Branch Test Map: `protectionLivenessAt`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/console/protection_liveness.go:53`: re-evaluate a read marker at response time and never upgrade a stopped read after wall-clock rollback | `TestA111PositionManagementNeverResurrectsAStoppedMarkerAfterClockRollback` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
