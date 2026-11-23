# Branch Test Map: `attestationWatch.eventFor`

- 편집 후(GREEN). 커버: `../../implementation/after-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA196RefusalTellsTheConsequenceFirst` · `TestA196ScopeLossWordingFollowsTheMissingKind` · `TestA196ImminentProofNoticeTellsTheProofDeadline`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@405 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 405.32,407.3 count 1 |
| B2 | if@410 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 410.27,412.3 count 1 |
| B3 | switch@430 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 431.34,432.38 count 1 |
| B4 | case@431 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 431.34,432.38 count 1 |
| B5 | if@432 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 432.38,436.4 count 1 |
| B6 | else@436 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 436.9,440.4 count 1 |
| B7 | case@441 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 441.25,444.130 count 1 |
| B8 | case@445 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 445.20,447.141 count 1 |
| B9 | case@448 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 448.22,452.31 count 1 |
| B10 | range@452 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 452.31,453.27 count 1 |
| B11 | if@453 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 453.27,455.5 count 1 |
| B12 | if@457 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 457.34,461.4 count 1 |
| B13 | else@461 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 461.9,464.21 count 1 |
| B14 | if@464 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 464.21,466.5 count 1 |
| B15 | case@469 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 469.23,471.157 count 1 |
| B16 | case@472 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 472.28,475.86 count 1 |
| B17 | case@476 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 476.20,481.85 count 1 |
| B18 | case@482 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 482.10,484.120 count 1 |
| B19 | if@488 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 488.43,489.18 count 1 |
| B20 | switch@489 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 490.54,491.100 count 1 |
| B21 | case@490 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 490.54,491.100 count 1 |
| B22 | case@492 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 492.11,493.139 count 1 |
| B23 | if@500 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 500.27,502.3 count 1 |
| B24 | if@503 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 503.56,505.3 count 1 |
| B25 | if@506 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 506.24,508.3 count 1 |
| B26 | if@509 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 509.25,511.3 count 1 |
| B27 | if@512 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 512.26,515.3 count 1 |
