# Branch Test Map: `attestationWatch.prepare`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA135WatchAnnouncesEachStageOnce` · `TestA196ScopeLossOncePerDayAcrossReissues`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@311 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 311.35,313.20 count 1 |
| B2 | if@313 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 313.20,317.5 count 0 |
| B3 | if@328 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 328.45,331.3 count 1 |
| B4 | switch@336 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 337.48,337.48 count 1 |
| B5 | case@337 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 337.48,337.48 count 1 |
| B6 | case@338 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 338.10,339.115 count 1 |
| B7 | if@341 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 341.26,343.3 count 1 |
| B8 | if@345 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 345.29,347.3 count 1 |
| B9 | if@349 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 349.17,351.3 count 1 |
