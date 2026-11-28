# Branch Test Map: `Service.LoadRawEngineAdoption`

편집 전 좌표(AST 4 분기, `internal/config/adoption_io.go` 47–73). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파일 없음 → 0 블록 | `TestRawLoadOfAMissingFileIsTheZeroBlock` | n/a | no |
| B2 | 읽기 실패 | 미측정 | n/a | no |
| B3 | 파싱 실패 | 미측정 | n/a | no |
| B4 | 거부 블록을 쓰인 그대로 | `TestRawLoadReturnsARefusedBlockAsWritten` | n/a | no |

a208: `TestA208ARefusedBlockClearsFloorOnly`(원문 읽기는 거부 블록의 토글도 보존) · `TestA208FloorOnlyExitRoundTrips` — RED 관측 `analysis/tdd/red.txt`.
