# Branch Test Map: `verifyGate`

- 편집 전. 커버: `../../implementation/base-cover-engine.out`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: `TestGateOnRefusals` · `TestNothingElseRefusesTheOperatorConfiguration` · `TestRefusedStartIsAudited`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@498 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 498.27,500.3 count 1 |
| B2 | if@506 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 506.49,508.3 count 1 |
| B3 | if@513 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 513.58,515.3 count 1 |
| B4 | if@519 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 519.16,523.3 count 1 |
| B5 | if@531 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 531.22,533.3 count 0 |
| B6 | if@536 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 536.79,538.3 count 1 |
| B7 | if@542 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 542.75,544.3 count 1 |
| B8 | if@546 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 546.76,548.3 count 1 |
| B9 | if@550 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 550.80,552.3 count 1 |
| B10 | if@555 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 555.52,557.3 count 0 |
| B11 | if@558 | 선택 집합(머리) | 해당 없음(기존 분기) | no — 558.66,560.3 count 0 |
