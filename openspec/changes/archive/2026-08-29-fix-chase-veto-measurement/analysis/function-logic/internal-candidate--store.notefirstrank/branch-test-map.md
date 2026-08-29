# Branch Test Map: `Store.NoteFirstRank`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 인자 검증이 트랜잭션 이전에 끝난다 | `TestARankOfZeroIsNotAFirstSighting` | — (기존 동작) | yes |
| B2 | 0·음수 rank/total | `TestARankOfZeroIsNotAFirstSighting` | — (기존 동작) | yes |
| B3 | 목록보다 큰 rank | `TestARankOfZeroIsNotAFirstSighting`(151/150) | — (기존 동작) | yes |
| B4 | instant 없는 위치 | `TestARankOfZeroIsNotAFirstSighting` | — (기존 동작) | yes |
| B5 | 요청 행 수가 음수 | `TestANegativeRequestedCountIsRefusedByTheFirstRankWrite` | yes (가드 무력화 시 실패 확인) | yes |
| B6 | source 없는 위치 | `TestARankOfZeroIsNotAFirstSighting` | — (기존 동작) | yes |
| B7 | BeginTx 실패 | **커버 없음** (I/O) | no | no |
| B8 | 기존 행을 자격 칼럼까지 읽는다 | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext` | yes | yes |
| B9 | 후보가 없는 심볼 | `TestARankOfZeroIsNotAFirstSighting`(KR:999999) | — (기존 동작) | yes |
| B10 | SELECT 실패 | **커버 없음** (I/O) | no | no |
| B11 | 이미 있는 위치는 덮이지 않고, 자격 칼럼도 나중 읽기로 채워지지 않는다 | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext` · `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan` | yes | yes |
| B12 | 읽을 수 없는 first_seen_at | **커버 없음** — 손상된 행 | no | no |
| B13 | 동일성 창 밖 읽기는 조용히 저장되지 않는다 | `TestARankFromOutsideTheIdentityWindowIsNotStored` | — (기존 동작) | yes |
| B14 | UPDATE 실패 | **커버 없음** (I/O) | no | no |
| B15 | Commit 실패 | **커버 없음** (I/O) | no | no |

**정직한 커버리지 기록**: B7·B10·B12·B14·B15는 I/O·손상 경로이고 주입 장치가 없다.
이 change가 만든 분기는 B5 하나이며 커버되어 있다(이번에 추가).
