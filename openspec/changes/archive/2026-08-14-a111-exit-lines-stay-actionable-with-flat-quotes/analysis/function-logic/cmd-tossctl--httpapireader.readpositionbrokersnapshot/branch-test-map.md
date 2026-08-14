# Branch Test Map: `httpAPIReader.readPositionBrokerSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `cmd/tossctl/httpapi_reader.go:108`: perform only the rate-limited broker/account read and clone cached rows | `TestA111CachedPositionsObserveRealEngineStopAndResumeWithoutContamination` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `cmd/tossctl/httpapi_reader.go:112`: perform only the rate-limited broker/account read and clone cached rows | `TestA111CachedPositionsObserveRealEngineStopAndResumeWithoutContamination` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `cmd/tossctl/httpapi_reader.go:116`: perform only the rate-limited broker/account read and clone cached rows | `TestA111CachedPositionsObserveRealEngineStopAndResumeWithoutContamination` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
