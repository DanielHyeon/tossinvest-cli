# Branch Test Map: `NewContext`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestGateOnRefusals` · `TestNothingElseRefusesTheOperatorConfiguration` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@453 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 453.16,455.3 count 0 |
| B2 | if@460 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 460.16,462.3 count 0 |
| B3 | if@464 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 464.16,466.3 count 1 |
| B4 | if@478 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 478.27,480.3 count 0 |
| B5 | if@486 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 486.2,487.71 count 1 |
| B6 | if@501 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 501.16,503.3 count 0 |
| B7 | if@508 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 508.16,510.3 count 0 |
| B8 | if@517 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 517.82,520.3 count 0 |
| B9 | if@524 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 524.31,526.17 count 0 |
| B10 | if@526 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 526.17,529.4 count 0 |
| B11 | if@545 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 545.16,548.3 count 0 |
| B12 | if@560 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 560.72,562.21 count 1 |
| B13 | if@562 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 562.21,564.4 count 1 |
| B14 | if@568 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 568.17,572.4 count 0 |
| B15 | if@589 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 589.16,592.3 count 1 |
| B16 | if@593 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 593.32,595.3 count 1 |
| B17 | if@598 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 598.16,601.3 count 0 |
