# Branch Test Map: `Console.readProtectionMarker`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/console/protection_liveness.go:42`: return unwired without filesystem access or one enginelock status read bounded by caller time | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
