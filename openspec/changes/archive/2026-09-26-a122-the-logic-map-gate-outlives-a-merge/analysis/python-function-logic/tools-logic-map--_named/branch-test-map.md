# Branch Test Map: `_named` (Python, 보수 — 새 함수)

## 보수 (독립 리뷰 둘, 2026-09-26)

짝은 변이로 잡았다(`75_mut.py` 창 `246:253` · `233:234`, 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 416`; 사본은 최종 파일과 주석 한 덩이만 다르고 AST 가 같다 — 실측).

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 원장에 적는다 | MR4 | `test_an_archived_copy_that_appears_while_judged_asks_for_a_rerun` |
| B1 실패면 올린다 | (변이 안 돌림) | `resolve_referenced_change` 가 `FileNotFoundError` · `NotADirectoryError` 만 "아카이브 없음" 으로 받는다 — 기존 해소기 시험 |
