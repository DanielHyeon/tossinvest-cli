# Branch Test Map: `TestIndex.__init__` (Python, 7.5.9 — 새 메서드)

## 보수 (독립 리뷰 둘, 2026-09-26)

짝은 변이로 잡았다(`75_mut.py` 창 `246:253` · `233:234`, 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 416`; 사본은 최종 파일과 주석 한 덩이만 다르고 AST 가 같다 — 실측).

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| (분기 없음) `files` 를 붙인다 | — | `resolve_test_file` 의 모든 시험이 `index.files` 를 쓴다 |
