# Branch Test Map: `Service.SaveEngineAdoption`

편집 뒤 좌표(a208 3.3 재추출 · AST 9 분기, `internal/config/adoption_io.go` 81–127). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 거부될 블록 → 쓰지 않음 | `TestSaveRefusesAnInvalidBlock` | n/a | yes — 선택 시험 통과(3.1) |
| B2 | 잠금 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B3 | ReadFile 결과 분기 | `TestSaveReplacesOnlyTheAdoptionBlock` | n/a | yes — 선택 시험 통과(3.1) |
| B4 | 파일 없음 → 뼈대 | `TestSaveCreatesASkeletonOnlyWhenTheFileIsAbsent` | n/a | yes — 선택 시험 통과(3.1) |
| B5 | 뼈대 직렬화 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B6 | 읽기 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B7 | 깨진 JSON → 거절 | `TestSaveRefusesACorruptFile` | n/a | yes — 선택 시험 통과(3.1) |
| B8 | 블록 직렬화 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B9 | splice 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |

a208: `TestA208FloorOnlyExitRoundTrips`(참 보존 · 거짓이면 키 없음) — RED 관측 `analysis/tdd/red.txt`.
