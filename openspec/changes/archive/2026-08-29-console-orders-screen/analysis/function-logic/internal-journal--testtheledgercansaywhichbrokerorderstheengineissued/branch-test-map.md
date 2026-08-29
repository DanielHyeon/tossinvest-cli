# Branch Test Map: `TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽기가 오류 없이 돌아온다 | 자체 실행 | yes (`BrokerOrderIDs` 부재로 컴파일 실패) | yes |
| B2 | 정렬 + 중복 제거 + 빈 id 제외가 동시에 성립한다 | 자체 실행(픽스처: `ord-b`, `ord-a`, `ord-a`, `''`) | yes | yes |
