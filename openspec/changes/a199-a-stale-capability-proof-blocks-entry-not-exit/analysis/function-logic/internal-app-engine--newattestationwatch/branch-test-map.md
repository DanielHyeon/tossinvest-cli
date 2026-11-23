# Branch Test Map: `newAttestationWatch`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA135ProductionExecutorShape`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@231 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 231.14,233.3 count 0 |
| B2 | if@234 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 234.20,236.3 count 0 |
| B3 | if@244 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 244.23,246.3 count 0 |
