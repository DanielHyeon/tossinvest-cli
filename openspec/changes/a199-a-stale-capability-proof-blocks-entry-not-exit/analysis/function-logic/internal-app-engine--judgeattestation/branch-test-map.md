# Branch Test Map: `judgeAttestation`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA135StageBoundaries` · `TestA135ScopeLossIsCritical` · `TestA135UnusableAttestationsAreCritical` · `TestA196DeadlineBoundaries`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@101 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 101.20,103.3 count 1 |
| B2 | if@106 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 106.62,107.10 count 1 |
| B3 | switch@107 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 108.42,109.26 count 1 |
| B4 | case@108 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 108.42,109.26 count 1 |
| B5 | case@110 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 110.54,112.46 count 1 |
| B6 | case@113 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 113.50,114.34 count 1 |
| B7 | case@115 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 115.45,116.29 count 1 |
| B8 | case@117 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 117.11,119.29 count 0 |
| B9 | if@127 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 127.20,132.3 count 1 |
| B10 | switch@135 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 136.42,137.25 count 1 |
| B11 | case@136 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 136.42,137.25 count 1 |
| B12 | case@138 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 138.44,139.24 count 1 |
| B13 | case@140 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 140.10,141.26 count 1 |
