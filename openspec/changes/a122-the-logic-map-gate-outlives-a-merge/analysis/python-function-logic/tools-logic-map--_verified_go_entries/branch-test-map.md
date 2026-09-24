# Branch Test Map: `_verified_go_entries` (Python, a122 task 7.5.34 — 새 함수)

변이 로그(`75_mut.py`, 창 `143:190`, 대조군 양끝 GREEN `Ran 372` — 하네스 사본은 `docs/WORKFLOW.md` 가 없어 skip 2)에서 옮겼다. 변이마다 값 하나. 이 함수는 변이 판 뒤에 바뀌지 않았다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B2 `ls-tree` 실패 | `AJ12_a_failed_ls_tree_is_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B5·B6 레코드 모양 | `AJ13_a_short_ls_tree_record_is_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B7 트리만 묻는다(gitlink 제외) | `AJ14_a_gitlink_is_asked_as_a_tree` | `test_a_submodule_in_the_base_is_not_refused` · `test_a_deleted_go_gitlink_is_refused_by_name` |
| B9·B10 커밋의 `tree` 줄 | `AJ15_the_commit_line_is_not_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B12·B13 목록이 빠뜨린 트리 | `AJ16_the_walk_trusts_the_listing` | `test_every_git_fault_of_the_comparison_is_named` |
| B15·B16 잘린 트리 항목 | `AJ17_a_short_tree_entry_is_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B17 하위 트리를 걷는다 | `AJ18_subtrees_are_not_walked` | 16 시험(첫 셋: `test_a_failure_prints_the_comparison_window` · `test_a_fault_while_asking_the_recorder_does_not_become_advice` · `test_a_forged_base_object_is_refused`) |
| B18 `.go` 만 | `AJ19_the_base_keeps_every_file` | 90 시험 |

B1 · B3 · B4 · B8 · B11 · B14(순회)는 따로 잴 값이 없다 — 빠지면 base 가 비어 모든 판정이 바뀐다.
