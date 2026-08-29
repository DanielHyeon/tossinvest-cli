# Branch Test Map: `TestARankFromOutsideTheIdentityWindowIsNotStored`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 승격 | 자체 실행 | yes (컴파일) | yes |
| B2 | 창 밖은 오류가 아니다 | 자체 실행 | yes | yes |
| B3 | 창 밖은 아무것도 돌려주지 않는다 | 자체 실행 | yes | yes |
| B4 | 창 밖은 쓰이지 않는다 | 자체 실행 | yes | yes |
| B5 | 창 안은 오류가 아니다 | 자체 실행 | yes | yes |
| B6 | 창 안은 저장된다 | 자체 실행 | yes | yes |
