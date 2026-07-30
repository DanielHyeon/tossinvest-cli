# Branch Test Map: `Store.Observations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | since가 있는 읽기와 없는 읽기 | `TestPruningRawObservationsLeavesTheFirstRankToo` · `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows` | — (기존 동작) | yes |
| B2 | 쿼리 실패 | **커버 없음** (I/O) | no | no |
| B3 | 새 두 칼럼을 실어 행을 돌려준다 | `TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows` · `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows` | yes (칼럼 수 불일치면 컴파일은 되고 Scan이 실패) | yes |
| B4 | 읽을 수 없는 instant | **커버 없음** | no | no |
| B5 | 반복 중 오류 | **커버 없음** (I/O) | no | no |

**정직한 커버리지 기록**: B2·B5는 SQLite I/O 실패이고 주입 장치가 없다. 이 change가 만든
분기가 아니다.
