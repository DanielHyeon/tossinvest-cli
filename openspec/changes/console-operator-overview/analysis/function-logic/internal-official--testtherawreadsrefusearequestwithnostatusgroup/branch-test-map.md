# Branch Test Map: `TestTheRawReadsRefuseARequestWithNoStatusGroup`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 토큰 요청은 세지 않는다 | 자체 실행 | yes (가드 부재 시 실요청이 카운트돼 B6이 실패) | yes |
| B2 | 두 메서드 × empty/blank 4케이스 | 자체 실행 | yes | yes |
| B3 | 타입이 있는 오류다 | 자체 실행 | yes | yes |
| B4 | 오류가 파라미터 이름을 댄다 | 자체 실행 | yes | yes |
| B5 | 오류가 `OPEN`을 댄다 | 자체 실행 | yes | yes |
| B6 | 거부된 요청이 브로커에 닿지 않는다(요청 수 0) | 자체 실행 | yes | yes |
| B7 | 호출자 실수를 fallback 신호로 보고하지 않는다 | 자체 실행 | yes | yes |

변이 검증(issues.md I-7): `OrdersRaw`/`ConditionalOrdersRaw`의 가드 조건을 항상 거짓이 되는
값으로 바꾸자 이 테스트가 실패했다.
