# Branch Test Map: `receiverIsConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 수신자 없는 함수 | 패키지의 최상위 함수 전부 | — | yes |
| B2 | 포인터 수신자를 벗긴다 | `SessionToken`·`Handler`·`Addr`·`URL`·`Serve` 등 exported 메서드 | — | yes |
