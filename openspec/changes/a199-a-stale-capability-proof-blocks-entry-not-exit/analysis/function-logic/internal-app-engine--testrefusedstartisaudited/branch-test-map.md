# Branch Test Map: `TestRefusedStartIsAudited`

- 편집 전. 이 시험은 base 에서 rc 0(`analysis/implementation/base-selected-engine.txt`(rc 0, `-run` 목록은 그 파일 머리가 아니라 review.md 1.9 에)). 시험 함수라 커버 대상이 아님 — 분기는 시험 자신의 사례 · 단언 갈래.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if@656 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 시험 rc 0(base) |
| B2 | if@662 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 시험 rc 0(base) |
| B3 | if@665 | 선택 집합(머리) | 해당 없음(기존 분기) | yes — 시험 rc 0(base) |
