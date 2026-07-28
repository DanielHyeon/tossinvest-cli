# Branch Test Map: `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v2 저장소가 열린다 | 자체 실행 | — (기존 동작) | yes |
| B2 | 칼럼 목록 | 자체 실행 | — (기존 동작) | yes |
| B3 | 기대 칼럼 존재 | 자체 실행 | — (기존 동작) | yes |
| B4 | 버전 읽기 | 자체 실행 | — (기존 동작) | yes |
| B5 | 버전이 현재 | 자체 실행 | yes (SchemaVersion 3→4) | yes |
| B6 | 후보 보존 | 자체 실행 | — (기존 동작) | yes |
| B7 | 생애 instant 보존 | 자체 실행 | — (기존 동작) | yes |
| B8 | provenance 보존 | 자체 실행 | — (기존 동작) | yes |
| B9 | baseline 읽기 | 자체 실행 | — (기존 동작) | yes |
| B10 | baseline 보존 | 자체 실행 | — (기존 동작) | yes |
| B11 | 관측 읽기 | 자체 실행 | — (기존 동작) | yes |
| B12 | 관측 2행 보존 | 자체 실행 | — (기존 동작) | yes |
| B13 | 최초 순위 읽기 | 자체 실행 | — (기존 동작) | yes |
| B14 | 백필된 위치 | 자체 실행 | — (기존 동작) | yes |
| B15 | **마이그레이션된 위치는 측정되지 않는다** | 자체 실행 | yes (전에는 92%로 측정 가능을 단언했다) | yes |
| B16 | 사유가 `NEW_ENTRANT_UNKNOWN` | 자체 실행 | yes | yes |
| B17 | 거부돼도 읽은 것을 보고한다 | 자체 실행 | yes | yes |
| B18 | gap 행 후보 | 자체 실행 | — (기존 동작) | yes |
| B19 | gap 행은 백필되지 않는다 | 자체 실행 | — (기존 동작) | yes |
| B20 | 닫기 | 자체 실행 | — (기존 동작) | yes |
| B21 | 재오픈 | 자체 실행 | — (기존 동작) | yes |
| B22 | 재오픈 후 위치 보존 | 자체 실행 | — (기존 동작) | yes |

회복 경로("생명이 끝나면 측정 가능해진다")는 이 테스트가 아니라
`TestAMigratedCandidateBecomesMeasurableWhenItsLifeEnds`가 잡고, 반대 방향("나중 스캔이
채워 넣지 않는다")은 `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan`이
잡는다. 세 테스트가 함께 정정된 주석의 세 문장을 각각 지지한다.
