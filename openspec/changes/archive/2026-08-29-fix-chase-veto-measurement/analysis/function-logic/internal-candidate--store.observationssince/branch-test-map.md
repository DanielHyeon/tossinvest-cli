# Branch Test Map: `Store.ObservationsSince`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 쿼리 실패 | **커버 없음** (I/O) | no | no |
| B2 | 시장 전체의 창 안 관측을 돌려준다 | `TestASessionStartDoesNotStampThePanelAsSeenLate` · `TestATruncatedReadingReachesTheVerdictAsTruncated` | yes | yes |
| B3 | 읽을 수 없는 instant | **커버 없음** | no | no |
| B4 | 반복 중 오류 | **커버 없음** (I/O) | no | no |

**정직한 커버리지 기록**: B1·B4는 I/O 실패 경로로 커버되지 않는다.
