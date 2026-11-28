# Branch Test Map: `Service.SaveEngineAdoption`

편집 전 좌표(AST 9 분기, `internal/config/adoption_io.go` 80–126). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 거부될 블록 → 쓰지 않음 | `TestSaveRefusesAnInvalidBlock` | n/a | no |
| B2 | 잠금 실패 | 미측정 | n/a | no |
| B3 | ReadFile 결과 분기 | `TestSaveReplacesOnlyTheAdoptionBlock` | n/a | no |
| B4 | 파일 없음 → 뼈대 | `TestSaveCreatesASkeletonOnlyWhenTheFileIsAbsent` | n/a | no |
| B5 | 뼈대 직렬화 실패 | 미측정 | n/a | no |
| B6 | 읽기 실패 | 미측정 | n/a | no |
| B7 | 깨진 JSON → 거절 | `TestSaveRefusesACorruptFile` | n/a | no |
| B8 | 블록 직렬화 실패 | 미측정 | n/a | no |
| B9 | splice 실패 | 미측정 | n/a | no |

a208: `TestA208FloorOnlyExitRoundTrips`(참 보존 · 거짓이면 키 없음) — RED 관측 `analysis/tdd/red.txt`.
