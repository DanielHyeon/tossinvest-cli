# Branch Test Map: `mergeAdoption`

편집 뒤 좌표(a208 3.3 재추출 · AST 3 분기, `internal/config/engine.go` 299–317). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 블록 없음 → 기본값 | `TestAdoptionDefaultsOff` | n/a | yes — 선택 시험 통과(3.1) |
| B2 | enabled 명시 | `TestAdoptionBlockParsed` | n/a | yes — 선택 시험 통과(3.1) |
| B3 | 범위 밖 손절폭 → 전면 0 | `TestAdoptionRefusesAStopFractionOutsideTheBand` | n/a | yes — 선택 시험 통과(3.1) |

a208: `TestA208FloorOnlyExitRoundTrips`(참 보존 · 키 없음 거짓) · `TestA208ARefusedBlockClearsFloorOnly`(B3 가 토글도 지움) — RED 관측 `analysis/tdd/red.txt`.
