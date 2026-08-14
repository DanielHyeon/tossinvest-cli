# Branch Test Map: `httpAPIReader.exitResponseAuthority`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `cmd/tossctl/httpapi_reader.go:324`: read the marker once between pre/post clocks, use the post clock for response age, and permit only downgrade of the read verdict | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `cmd/tossctl/httpapi_reader.go:327`: read the marker once between pre/post clocks, use the post clock for response age, and permit only downgrade of the read verdict | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `cmd/tossctl/httpapi_reader.go:333`: stopped marker read plus post-marker wall rollback remains stopped (downgrade-only) | `TestA111PostMarkerClockRollbackCannotResurrectAStoppedEngine` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B4 | `cmd/tossctl/httpapi_reader.go:335`: read the marker once between pre/post clocks, use the post clock for response age, and permit only downgrade of the read verdict | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
