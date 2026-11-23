# Branch Test Map: `Context.ExitObserver`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestA092ExitObserverGetsRecordOnlyAlertPaths` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@320 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 320.14,322.3 count 0 |
| B2 | if@323 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 323.28,325.3 count 0 |
| B3 | if@327 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 327.9,330.3 count 0 |
| B4 | if@344 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 344.23,346.3 count 1 |
| B5 | if@349 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 349.23,354.3 count 1 |
| B6 | if@361 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 361.23,363.3 count 1 |
