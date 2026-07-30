# Branch Test Map: `TestTheAdaptedOrderReadCannotTellAnAbsentPriceFromAZeroOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽기가 오류 없이 돌아온다 | 자체 실행 | yes (`ordersRawServer`/`ordersRawClient` 부재로 컴파일 실패) | yes |
| B2 | 픽스처 2건 | 자체 실행 | yes | yes |
| B3 | null 가격이 0으로 접힌다(`Orders`의 현재 행동) | 자체 실행 | yes | yes |
| B4 | null 체결이 0으로 접힌다 | 자체 실행 | yes | yes |
