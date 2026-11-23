# Branch Test Map: `newAttestationWatch`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA135ProductionExecutorShape`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@221 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 221.14,223.3 count 0 |
| B2 | if@224 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 224.20,226.3 count 0 |
| B3 | if@233 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 233.23,235.3 count 0 |
