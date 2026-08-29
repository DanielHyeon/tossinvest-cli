# Branch Test Map: `TestAnUnreportedBudgetStaysUnreported`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 접근자 없는 소스 생성 | 자체 실행 | yes (컴파일) | yes |
| B2 | 읽기 | 자체 실행 | yes | yes |
| B3 | 접근자 부재는 미보고 | 자체 실행 | — (기존 동작) | yes |
| B4 | 침묵 소스 생성 | 자체 실행 | yes | yes |
| B5 | 읽기 | 자체 실행 | yes | yes |
| B6 | 헤더 부재는 미보고 | 자체 실행 | — (기존 동작) | yes |
