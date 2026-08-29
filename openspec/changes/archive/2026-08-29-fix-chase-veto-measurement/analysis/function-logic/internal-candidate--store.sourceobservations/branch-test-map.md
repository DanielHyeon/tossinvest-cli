# Branch Test Map: `Store.SourceObservations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈/공백 source는 거부 | **커버 없음** | no | no |
| B2 | since 유무 | `TestOneSourcesReadingsCanBeReadWithoutTheOthers` | — (기존 동작) | yes |
| B3 | 쿼리 실패 | **커버 없음** (I/O) | no | no |
| B4 | 새 두 칼럼이 시리즈까지 실려 온다 | `TestOneSourcesReadingsCanBeReadWithoutTheOthers`(왕복) · `TestAStoreAtSchemaThreeGainsTheFourColumnsAndTheirConstraints`(마이그레이션된 행은 미상) | yes | yes |
| B5 | 읽을 수 없는 instant | **커버 없음** | no | no |
| B6 | 반복 중 오류 | **커버 없음** (I/O) | no | no |

**정직한 커버리지 기록**: B3·B6은 I/O 실패 경로로 커버되지 않는다.
`RankChange`가 `NewlyListed`를 읽으므로 B4가 새 칼럼을 싣는다는 것은 저장→읽기 왕복으로
`internal/candidate`의 시리즈 테스트가 간접 확인한다.
