# Branch Test Map: `consoleOrderRecords`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `price:null`, `execution:null`인 미체결 주문 1건이 빈 문자열로 도착 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately` (마지막 단언) | yes (변환 경로에서 "0") | yes |
