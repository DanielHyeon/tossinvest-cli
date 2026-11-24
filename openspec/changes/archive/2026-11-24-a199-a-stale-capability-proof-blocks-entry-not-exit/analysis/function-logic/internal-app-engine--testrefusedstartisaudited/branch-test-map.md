# Branch Test Map: `TestRefusedStartIsAudited`

- 편집 후(GREEN). 이 시험은 GREEN 뒤 선택 실행 에서 rc 0. 시험 함수라 커버 대상이 아님 — 분기는 시험 자신의 사례 · 단언 갈래.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@634 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 시험 rc 0(GREEN 뒤 선택 실행) |
| B2 | if@640 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 시험 rc 0(GREEN 뒤 선택 실행) |
| B3 | if@644 | 선택 집합(머리) | red-1 · red-2(`../../tdd/red.txt`) — 분기별 귀속은 변이 원장 | yes — 시험 rc 0(GREEN 뒤 선택 실행) |
