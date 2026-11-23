# Branch Test Map: `judgeAttestation`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA135StageBoundaries` · `TestA135ScopeLossIsCritical` · `TestA135UnusableAttestationsAreCritical` · `TestA196DeadlineBoundaries`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@103 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 103.20,105.3 count 1 |
| B2 | if@109 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 109.18,110.20 count 1 |
| B3 | if@110 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 110.20,112.4 count 1 |
| B4 | else@112 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 112.9,115.4 count 1 |
| B5 | if@118 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 118.16,119.10 count 1 |
| B6 | switch@119 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 120.50,121.34 count 1 |
| B7 | case@120 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 120.50,121.34 count 1 |
| B8 | case@122 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 122.45,123.29 count 1 |
| B9 | case@124 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 124.11,126.29 count 0 |
| B10 | if@134 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 134.20,139.3 count 1 |
| B11 | switch@142 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 143.42,144.25 count 1 |
| B12 | case@143 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 143.42,144.25 count 1 |
| B13 | case@145 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 145.44,146.24 count 1 |
| B14 | case@147 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 147.10,148.26 count 1 |
