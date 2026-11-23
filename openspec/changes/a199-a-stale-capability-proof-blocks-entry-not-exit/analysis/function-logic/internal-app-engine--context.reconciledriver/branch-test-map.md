# Branch Test Map: `Context.ReconcileDriver`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestTheDriverRefusesAnUnverifiedGate`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@365 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 365.14,367.3 count 0 |
| B2 | if@368 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 368.28,370.3 count 1 |
| B3 | if@382 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 382.24,384.3 count 0 |
| B4 | if@385 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 385.56,387.3 count 0 |
| B5 | if@388 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 388.45,390.3 count 0 |
| B6 | if@391 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 391.21,393.3 count 0 |
