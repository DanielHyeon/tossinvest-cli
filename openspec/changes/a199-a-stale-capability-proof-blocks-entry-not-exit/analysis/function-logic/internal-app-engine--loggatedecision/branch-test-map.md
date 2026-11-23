# Branch Test Map: `logGateDecision`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestGateDecisionIsLoggedStructurally`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@479 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 479.61,483.3 count 1 |
| B2 | if@484 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 484.20,487.3 count 1 |
