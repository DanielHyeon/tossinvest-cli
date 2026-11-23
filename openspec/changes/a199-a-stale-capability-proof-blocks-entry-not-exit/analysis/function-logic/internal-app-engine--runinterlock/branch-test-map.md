# Branch Test Map: `runInterlock`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestGateOnRefusals` · `TestGateOffStartsAndDoesNoGateWork` · `TestGateDecisionIsLoggedStructurally`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@393 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 393.19,398.3 count 1 |
| B2 | if@400 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 400.51,404.3 count 1 |
| B3 | if@410 | 선택 집합(머리) | 해당 없음(기존 분기) | 미측정(블록 없음) |
