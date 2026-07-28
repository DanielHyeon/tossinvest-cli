# Branch Test Map: `TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽기가 오류 없이 돌아온다 | 자체 실행 | yes (`OrdersRaw` 부재로 컴파일 실패) | yes |
| B2 | 픽스처 2건 | 자체 실행 | yes | yes |
| B3 | 부재/실제값 기대표 순회 | 자체 실행 | yes | yes |
| B4 | null → `""`, 진짜 0 → `"0"` | 자체 실행 | yes | yes |
| B5 | `status`가 브로커 어휘 그대로 | 자체 실행 | yes | yes |
| B6 | 주문 id가 보존된다 | 자체 실행 | yes | yes |
