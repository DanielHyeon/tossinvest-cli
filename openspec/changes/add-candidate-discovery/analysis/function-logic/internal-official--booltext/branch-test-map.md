# Branch Test Map: `boolText`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `true`면 `"true"`, 아니면 `"false"` — 두 값 모두 호출된다 | `TestTheRawOrderReadReportsThatThePageWasTruncated`(true 경로) + `TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne`·`…DerivesTheMarket…`·`TestTheAdaptedOrderRead…`(false 경로) | yes (헬퍼 부재로 컴파일 실패) | yes |
