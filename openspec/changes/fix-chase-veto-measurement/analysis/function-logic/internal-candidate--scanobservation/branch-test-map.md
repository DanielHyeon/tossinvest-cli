# Branch Test Map: `scanObservation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 칼럼 수·순서 불일치 | 세 SELECT를 쓰는 모든 관측 테스트가 즉시 실패한다(전용 테스트는 없음) | yes (칼럼을 하나 빼면 전부 실패) | yes |
| B2 | 읽을 수 없는 instant | `store_test.go`의 stamp 케이스 | — (기존 동작) | yes |
| B3 | 저장된 요청 행 수가 왕복한다 | `TestATruncatedReadingReachesTheVerdictAsTruncated` · `TestAWholeReadingOfTheSameLengthIsMeasured` · `TestAStoreAtSchemaThreeGainsTheFourColumnsAndTheirConstraints` | yes | yes |

schema-1 `newly_listed`를 **읽지 않는다**는 것 자체를 잡는 테스트는 없다 —
`TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`가 백필 없는 마이그레이션 뒤
`NEW_ENTRANT_UNKNOWN`을 요구하므로 간접적으로만 선다(그 칼럼을 읽으면 `no`가 되어 측정
가능해지고 그 테스트가 실패한다). 직접적인 단위 테스트는 없다.
