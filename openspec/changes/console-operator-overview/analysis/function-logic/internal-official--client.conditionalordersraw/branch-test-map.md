# Branch Test Map: `Client.ConditionalOrdersRaw`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈/공백 status는 요청 전에 거부되고 `ShouldFallback`이 false | `TestTheRawReadsRefuseARequestWithNoStatusGroup`(`ConditionalOrdersRaw/empty`, `/blank`; 서버 요청 수 0) | yes (가드를 항상 거짓인 조건으로 바꾸자 테스트가 물었다 — issues.md I-7) | yes |
| B2 | 주어진 그룹이 그대로 실린다 | `TestTheRawConditionalReadKeepsAnAbsentValueApartFromAZeroOne`(`OPEN`) | yes | yes |
| B3 | `symbol` 생략 시 부재 | 동 테스트(빈 값 전달, 서버가 경로만 처리) | yes | yes |
| B4 | `cursor` 생략 시 부재 | 동상 | yes | yes |
| B5 | `limit<=0`이면 부재 | 동상(`0` 전달) | yes | yes |
| B6 | 전송·인증·429·5xx가 분류된 오류로 올라온다 | `orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead`(같은 `send` 경로) | — | yes |
| B7 | 각 조건주문이 원문 문자열로 옮겨지고 페이지 경계가 보존된다 | `TestTheRawConditionalReadKeepsAnAbsentValueApartFromAZeroOne`(`orderPrice` null → `""`, `hasNext`/`nextCursor` 단언) | yes | yes |
