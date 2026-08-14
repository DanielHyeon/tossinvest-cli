# Branch Test Map: `httpAPIReader.Positions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `cmd/tossctl/httpapi_reader.go:101`: use cacheAt only for broker-cache expiry and delegate all local response authority to projectPositions | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
