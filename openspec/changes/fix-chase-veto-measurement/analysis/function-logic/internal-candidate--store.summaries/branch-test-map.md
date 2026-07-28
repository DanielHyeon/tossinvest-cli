# Branch Test Map: `Store.Summaries`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 쿼리 실패 | **커버 없음** (I/O) | no | no |
| B2 | 후보마다 두 stored first를 싣는다 | `TestASessionStartDoesNotStampThePanelAsSeenLate` · `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan` | yes | yes |
| B3 | 칼럼 불일치 | 전 스캔 테스트가 즉시 실패(전용 테스트 없음) | yes | yes |
| B4 | 읽을 수 없는 first_seen_at | **커버 없음** | no | no |
| B5 | 읽을 수 없는 last_seen_at | **커버 없음** | no | no |
| B6 | 냉각된 후보 | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` | — (기존 동작) | yes |
| B7 | 읽을 수 없는 cooled_at | **커버 없음** | no | no |
| B8 | 손상된 baseline | **커버 없음** | no | no |
| B9 | 손상된 first_rank | **커버 없음** | no | no |
| B10 | 반복 중 오류 | **커버 없음** (I/O) | no | no |

**정직한 커버리지 기록**: B1·B4·B5·B7·B8·B9·B10은 손상된 행과 I/O 실패 경로이고 이
저장소에 주입 장치가 없다. 전부 이 change 이전부터 있던 분기다.
