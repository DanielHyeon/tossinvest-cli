# Branch Test Map: `TestARankedRowFromTheGapBetweenTwoLivesIsNotThisLifesFirstSighting`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 첫 생명 관측 | 자체 실행 | — (기존 동작) | yes |
| B2 | 첫 승격 | 자체 실행 | — (기존 동작) | yes |
| B3 | 냉각 | 자체 실행 | — (기존 동작) | yes |
| B4 | 첫 생명 최초 순위 | 자체 실행 | yes (시그니처 변경으로 컴파일 실패) | yes |
| B5 | gap 관측 | 자체 실행 | — (기존 동작) | yes |
| B6 | 재탄생 관측 | 자체 실행 | — (기존 동작) | yes |
| B7 | 재승격 | 자체 실행 | — (기존 동작) | yes |
| B8 | 새 `first_seen_at` | 자체 실행 | — (기존 동작) | yes |
| B9 | 새 생명 최초 순위 | 자체 실행 | yes (시그니처 변경으로 컴파일 실패) | yes |
| B10 | 최초 순위 읽기 | 자체 실행 | — (기존 동작) | yes |
| B11 | 저장된 위치가 새 생명의 것 | 자체 실행 | — (기존 동작) | yes |
| B12 | 후보 읽기 | 자체 실행 | — (기존 동작) | yes |
| B13 | 측정된다 | 자체 실행 | — (기존 동작) | yes |
| B14 | 4 of 150 | 자체 실행 | — (기존 동작) | yes |
| B15 | seen_late dangerous | 자체 실행 | — (기존 동작) | yes |
