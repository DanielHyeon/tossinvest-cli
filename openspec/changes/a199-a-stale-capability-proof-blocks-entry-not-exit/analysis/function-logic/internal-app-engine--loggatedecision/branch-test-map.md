# Branch Test Map: `logGateDecision`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestGateDecisionIsLoggedStructurally`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@496 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 496.67,500.3 count 1 |
| B2 | if@501 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 501.20,504.3 count 1 |
| B3 | if@505 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 505.21,510.3 count 1 |
