# Branch Test Map: `Service.LoadRawEngineAdoption`

편집 뒤 좌표(a208 3.3 재추출 · AST 4 분기, `internal/config/adoption_io.go` 47–74). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파일 없음 → 0 블록 | `TestRawLoadOfAMissingFileIsTheZeroBlock` | n/a | yes — 선택 시험 통과(3.1) |
| B2 | 읽기 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B3 | 파싱 실패 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B4 | 거부 블록을 쓰인 그대로 | `TestRawLoadReturnsARefusedBlockAsWritten` | n/a | yes — 선택 시험 통과(3.1) |

a208: `TestA208ARefusedBlockClearsFloorOnly`(원문 읽기는 거부 블록의 토글도 보존) · `TestA208FloorOnlyExitRoundTrips` — RED 관측 `analysis/tdd/red.txt`.
