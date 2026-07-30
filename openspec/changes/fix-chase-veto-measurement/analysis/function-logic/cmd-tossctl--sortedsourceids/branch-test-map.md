# Branch Test Map: `sortedSourceIDs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 두 소스가 결정적 순서로 렌더된다 | `TestTheJSONReportCarriesBothBlocks`(`got.Readings[0]`이 official trading value여야 한다) · `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived` | yes | yes |
