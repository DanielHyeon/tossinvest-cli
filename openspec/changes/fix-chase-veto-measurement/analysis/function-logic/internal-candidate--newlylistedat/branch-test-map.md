# Branch Test Map: `newlyListedAt`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 슬라이스 순회 — 대체된 경로는 `Store.FirstRank`의 칼럼 SELECT | `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan`(슬라이스가 아니라 칼럼에서 온다는 것을 5회 스캔으로 고정) | 삭제됨 | n/a |
| B2 | 4중 일치 — 대체된 경로는 `first.NewlyListed` 직접 읽기 | `TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate` · `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`(마이그레이션된 행이 `unknown`) | 삭제됨 | n/a |

이 함수 자신에 대한 테스트는 base에도 **없었다** — 그것이 10분 창 밖에서 항상 `false`를
돌려주고 있다는 사실이 몇 달 동안 아무것도 실패시키지 않은 이유의 일부다. 대체 경로에는
직접 테스트가 넷 있고 위에 적었다.
