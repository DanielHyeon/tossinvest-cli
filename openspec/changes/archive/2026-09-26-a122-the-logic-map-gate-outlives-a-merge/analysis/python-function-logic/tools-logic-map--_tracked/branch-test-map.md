# Branch Test Map: `_tracked` (Python, a122 task 7.5.9)

## task 7.5.9 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 결과를 원장에 적는다 | AA3(재조준 — 트리 순회 대신 추적 목록) | `test_a_test_file_that_appears_after_the_index_was_built_asks_for_a_rerun` |
| B1 실패면 올린다 | AM9(빈 목록) | `test_a_tracked_file_list_that_cannot_be_read_is_a_fault_not_an_empty_index` |
