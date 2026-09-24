# Branch Test Map: `_isolated_tree` (Python, a122 task 7.5.34 — 새 함수)

변이 로그(`75_mut.py`, 창 `143:190`, 대조군 양끝 GREEN `Ran 372` — 하네스 사본은 `docs/WORKFLOW.md` 가 없어 skip 2)에서 옮겼다. 변이마다 값 하나. 이 함수는 변이 판 뒤에 바뀌지 않았다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| `core.splitIndex=false` | `AJ31_the_split_index_leaks` | `test_the_snapshot_writes_nothing_into_the_repository_and_runs_nothing_it_configures` |
| `core.hooksPath=/dev/null` | `AJ32_the_hooks_run` | 같은 시험 |
| `write-tree` 도 고정을 받는다 | `AJ33_write_tree_takes_no_pins` | 같은 시험 |
| B2 `update-index` 실패 | `AJ34_a_failed_update_index_is_ignored` | `test_every_git_fault_of_the_comparison_is_named` |
| B3 `write-tree` 실패 | `AJ35_a_failed_write_tree_is_ignored` | `test_every_git_fault_of_the_comparison_is_named` |
| B4 `write-tree` 의 답 | `AJ36_the_write_tree_answer_is_not_read` | `test_every_git_fault_of_the_comparison_is_named` |
| `--missing-ok` | `AJ37_absent_blobs_stop_the_tree` | 121 시험 |
