# Branch Test Map: `insertAttemptWithBrokerOrder`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | INSERT 실패를 삼키지 않고 즉시 실패시킨다 | `TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued`가 이 헬퍼를 4회 호출 | yes (헬퍼 부재로 컴파일 실패) | yes |
