# Branch Test Map: `ordersRawClient`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) 계좌 seq가 고정된 클라이언트 하나 — 요청 수 단언이 계좌 해석에 오염되지 않는다 | `TestTheRawReadsRefuseARequestWithNoStatusGroup`(요청 0건), `TestBothOrderReadsSendTheGroupTheyWereGiven`(요청 2건) | yes (헬퍼 부재로 컴파일 실패) | yes |
