# Branch Test Map: `TestTheReadOnlyHandleHasNoWriteMethods`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `*ReadOnly`의 모든 exported 메서드를 순회한다 | 자체 실행 | yes (`BrokerOrderIDs` 추가 직후 allowlist 갱신 전 실패) | yes |
| B2 | allowlist에 없는 메서드를 보고한다 | 자체 실행 | yes (동상) | yes |
