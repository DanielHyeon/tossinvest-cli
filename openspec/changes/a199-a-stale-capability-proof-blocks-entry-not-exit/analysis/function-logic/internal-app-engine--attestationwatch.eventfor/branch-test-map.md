# Branch Test Map: `attestationWatch.eventFor`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA196RefusalTellsTheConsequenceFirst` · `TestA196ScopeLossWordingFollowsTheMissingKind` · `TestA196ImminentProofNoticeTellsTheProofDeadline`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@385 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 385.32,387.3 count 1 |
| B2 | if@390 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 390.27,392.3 count 1 |
| B3 | switch@407 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 408.34,409.38 count 1 |
| B4 | case@408 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 408.34,409.38 count 1 |
| B5 | if@409 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 409.38,413.4 count 1 |
| B6 | else@413 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 413.9,417.4 count 1 |
| B7 | case@418 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 418.25,421.130 count 1 |
| B8 | case@422 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 422.20,424.123 count 1 |
| B9 | case@425 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 425.22,429.31 count 1 |
| B10 | range@429 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 429.31,430.27 count 1 |
| B11 | if@430 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 430.27,432.5 count 1 |
| B12 | if@434 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 434.34,438.4 count 1 |
| B13 | else@438 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 438.9,441.21 count 1 |
| B14 | if@441 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 441.21,443.5 count 1 |
| B15 | case@446 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 446.23,448.157 count 1 |
| B16 | case@449 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 449.28,452.86 count 1 |
| B17 | case@453 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 453.10,455.120 count 1 |
| B18 | if@461 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 461.27,463.3 count 1 |
| B19 | if@464 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 464.56,466.3 count 1 |
| B20 | if@467 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 467.24,469.3 count 1 |
| B21 | if@470 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 470.25,472.3 count 1 |
| B22 | if@473 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 473.26,476.3 count 1 |
