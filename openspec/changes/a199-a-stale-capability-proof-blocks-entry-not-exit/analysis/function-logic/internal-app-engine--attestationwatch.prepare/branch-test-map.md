# Branch Test Map: `attestationWatch.prepare`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA135WatchAnnouncesEachStageOnce` · `TestA196ScopeLossOncePerDayAcrossReissues`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@322 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 322.35,324.20 count 1 |
| B2 | if@324 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | no — 324.20,328.5 count 0 |
| B3 | if@339 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 339.45,344.25 count 1 |
| B4 | range@344 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 344.25,345.54 count 1 |
| B5 | if@345 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 345.54,347.5 count 1 |
| B6 | switch@355 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 356.48,356.48 count 1 |
| B7 | case@356 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 356.48,356.48 count 1 |
| B8 | case@357 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 357.10,358.115 count 1 |
| B9 | if@360 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 360.26,362.3 count 1 |
| B10 | if@364 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 364.44,367.3 count 1 |
| B11 | if@369 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 369.17,371.3 count 1 |
