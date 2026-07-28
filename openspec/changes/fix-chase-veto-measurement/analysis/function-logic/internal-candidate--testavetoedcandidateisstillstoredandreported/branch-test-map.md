# Branch Test Map: `TestAVetoedCandidateIsStillStoredAndReported`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 관측 기록 | 자체 실행 | — (기존 동작) | yes |
| B2 | 승격 | 자체 실행 | — (기존 동작) | yes |
| B3 | 첫 가격 | 자체 실행 | — (기존 동작) | yes |
| B4 | 첫 순위 | 자체 실행 | yes (시그니처 변경으로 컴파일 실패) | yes |
| B5 | 후보 읽기 | 자체 실행 | — (기존 동작) | yes |
| B6 | baseline 읽기 | 자체 실행 | — (기존 동작) | yes |
| B7 | 최초 순위 읽기 | 자체 실행 | — (기존 동작) | yes |
| B8 | 관측 읽기 | 자체 실행 | — (기존 동작) | yes |
| B9 | veto 발화 | 자체 실행 | — (기존 동작) | yes |
| B10 | veto 후 후보 | 자체 실행 | — (기존 동작) | yes |
| B11 | `first_seen_at` 보존 | 자체 실행 | — (기존 동작) | yes |
| B12 | 관측 재읽기 | 자체 실행 | — (기존 동작) | yes |
| B13 | 관측 보존 | 자체 실행 | — (기존 동작) | yes |
| B14 | vetoed 1건 | 자체 실행 | — (기존 동작) | yes |
| B15 | total 1건 | 자체 실행 | — (기존 동작) | yes |
