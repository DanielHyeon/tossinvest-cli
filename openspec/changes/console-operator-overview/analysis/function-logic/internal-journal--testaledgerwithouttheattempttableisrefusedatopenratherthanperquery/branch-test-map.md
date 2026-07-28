# Branch Test Map: `TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파괴 문장 3개를 순서대로 실행한다 | 자체 실행 | yes (`readOnlyTables` 미등록 시 open이 성공해 B4에서 실패) | yes |
| B2 | 파괴 문장 실패를 삼키지 않는다 | 자체 실행 | — | yes |
| B3 | writer handle을 닫는다(WAL 잔여 회피) | 자체 실행 | — | yes |
| B4 | `OpenReadOnly`가 `ErrSchemaTooOld`로 거절한다 | 자체 실행 | yes (등록 전에는 open이 성공했다) | yes |
| B5 | 거절 문장이 `mutation_attempts`를 이름으로 부른다 | 자체 실행 | yes | yes |
