# Branch Test Map: `_worktree_entries` (Python, a122 task 7.5.34 — 새 함수)

변이 로그(`75_mut.py`, 창 `143:190`, 대조군 양끝 GREEN `Ran 372` — 하네스 사본은 `docs/WORKFLOW.md` 가 없어 skip 2)에서 옮겼다. 변이마다 값 하나. 이 함수는 변이 판 뒤에 바뀌지 않았다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 `ls-files` 실패 | `AJ20_a_failed_ls_files_is_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B4·B5 레코드 모양 | `AJ21_a_short_ls_files_record_is_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B6 `.go` 만(pathspec 없이) | `AJ22_the_worktree_keeps_every_file` | 9 시험(첫 셋: FIFO 둘 · `test_a_submodule_in_the_base_is_not_refused`) |
| 이름을 `os.fsdecode` 로 | `AJ23_an_odd_name_is_decoded_strictly` | `test_an_unchanged_file_with_a_name_that_is_not_utf8_is_not_refused` |
| B7 충돌 중 | `AJ24_an_unmerged_file_is_taken` | `test_the_snapshot_refuses_what_it_cannot_represent` |
| B8 일반 파일 모드가 아님 | `AJ25_a_symlink_is_taken` | 같은 시험 |
| B8 `100755` 는 일반 파일 | `AH14_an_executable_is_refused` | 69 시험(창 `96:143`) |
| B9 깔때기로 읽는다(원장) | `AJ26_reads_skip_the_ledger` | `test_the_snapshot_reads_through_the_funnel` |
| B11 skip-worktree + 없음 = sparse | `AJ27_a_sparse_file_is_a_deletion` | `test_a_file_sparse_checkout_left_out_is_not_a_deletion` · `test_a_sparse_entry_that_differs_from_the_base_is_judged_from_its_verified_blob` |
| B11 assume-unchanged + 없음 = 삭제 | `AJ28_assume_unchanged_keeps_the_index_blob` | `test_assume_unchanged_does_not_hide_a_deletion` |
| oid 가 git 의 blob id | `AJ29_the_digest_is_not_a_blob_id` | `test_an_untouched_worktree_is_an_empty_comparison` |
| `ls-files` 에 fsmonitor 고정 | `AJ30_the_ls_files_runs_the_fsmonitor` | `test_the_snapshot_writes_nothing_into_the_repository_and_runs_nothing_it_configures` |
| `SNAPSHOT_PINS` 자체 | `AH10_the_fsmonitor_runs` | 같은 시험(창 `96:143`) |

B2 · B3(`-z` 끝의 빈 조각)은 변이를 넣지 않았다 — 빼면 빈 조각이 B4·B5 의 결함이 되어 **모든** 판정이 멈춘다.
