# Branch Test Map: `TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v1 저장소가 열린다 | 자체 실행 | — (기존 동작) | yes |
| B2 | 칼럼 목록 | 자체 실행 | — (기존 동작) | yes |
| B3 | 기대 칼럼 존재 | 자체 실행 | — (기존 동작) | yes |
| B4 | 버전 읽기 | 자체 실행 | — (기존 동작) | yes |
| B5 | 버전이 현재 | 자체 실행 | — (기존 동작) | yes |
| B6 | 후보 보존 | 자체 실행 | — (기존 동작) | yes |
| B7 | `first_seen_at` 보존 | 자체 실행 | — (기존 동작) | yes |
| B8 | sources 보존 | 자체 실행 | — (기존 동작) | yes |
| B9 | provenance 보존 | 자체 실행 | — (기존 동작) | yes |
| B10 | 관측 읽기 | 자체 실행 | — (기존 동작) | yes |
| B11 | 관측 4행 보존 | 자체 실행 | — (기존 동작) | yes |
| B12 | baseline 읽기 | 자체 실행 | — (기존 동작) | yes |
| B13 | baseline 백필 | 자체 실행 | — (기존 동작) | yes |
| B14 | 나중 가격 기록 | 자체 실행 | — (기존 동작) | yes |
| B15 | baseline write-once | 자체 실행 | — (기존 동작) | yes |
| B16 | 최초 순위 읽기 | 자체 실행 | — (기존 동작) | yes |
| B17 | 최초 순위 백필 | 자체 실행 | — (기존 동작) | yes |
| B18 | 두 번째 후보 | 자체 실행 | — (기존 동작) | yes |
| B19 | 늦은 행은 백필되지 않는다 | 자체 실행 | — (기존 동작) | yes |
| B20 | 백필 없는 후보는 미측정 | 자체 실행 | — (기존 동작) | yes |
| B21 | 나중 순위 기록 | 자체 실행 | yes (시그니처 변경으로 컴파일 실패) | yes |
| B22 | 최초 순위 write-once | 자체 실행 | — (기존 동작) | yes |
| B23 | 닫기 | 자체 실행 | — (기존 동작) | yes |
| B24 | 재오픈 | 자체 실행 | — (기존 동작) | yes |
| B25 | 재오픈 후 baseline | 자체 실행 | — (기존 동작) | yes |

**정직한 커버리지 기록**: B2·B3의 칼럼 목록은 `first_rank_source`에서 멈춘다 —
schema-4의 네 칼럼(`newly_listed_state`, `rank_requested`, `first_rank_newly_listed`,
`first_rank_requested`)을 명명하지 않는다. 그 공백을 `schema_four_test.go`의 두 테스트가
메운다(v3 fixture + 칼럼·CHECK 고정).
