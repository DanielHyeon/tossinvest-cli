# Branch Test Map: `_worktree_snapshot` (Python, a122 task 7.5.25)

짝은 손으로 고르지 않았다 — 갈래마다 **그 갈래 하나만** 바꾸는 변이를 `75_mut.py` 에 넣고, 실제로 빨개진 시험을 변이 로그에서
옮겼다. 두 값을 한꺼번에 빼는 변이는 넣지 않았다(7.5.28 의 `AG2` 가 그렇게 해서 BTM 이 한쪽을 부풀렸다 — 재리뷰 F1).
창 `120:160` · `147:151` · `138:139`, 대조군 양끝 GREEN.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 `rev-parse` 실패 | `AH21_a_failed_rev_parse_is_read` | `test_every_git_fault_of_the_snapshot_is_named` |
| B2·B3 `rev-parse` 답이 모자람 | `AH22_a_short_rev_parse_is_read` | 같은 시험 |
| B4 `ls-files` 실패 | `AH23_a_failed_ls_files_is_read` | 같은 시험 |
| B7·B8 레코드 모양 | `AH24_a_short_record_is_read` | 같은 시험 |
| B9 충돌 중 | `AH12_an_unmerged_file_is_taken` | `test_the_snapshot_refuses_what_it_cannot_represent` |
| B10 일반 파일 모드가 아님 | `AH13_a_symlink_is_taken` | 같은 시험 |
| B10 `100755` 는 일반 파일이다 | `AH14_an_executable_is_refused` | 69 시험(대부분의 실저장소 픽스처) — 그중 `test_an_executable_flagged_file_is_still_seen` |
| B11 깔때기로 읽는다(원장) | `AH7_reads_skip_the_ledger` | `test_the_snapshot_reads_through_the_funnel` |
| B11 깔때기로 읽는다(원시 호출) | `(root / path).read_bytes()` — 따로 쟀다 | `test_no_read_primitive_lives_outside_the_funnels` · `test_the_snapshot_reads_through_the_funnel`; **FIFO 시험은 멈춘다**(깔때기가 막던 그 멈춤) |
| B12·B13 skip-worktree + 없음 = sparse | `AH5_a_sparse_file_is_a_deletion` | `test_a_file_sparse_checkout_left_out_is_not_a_deletion` |
| B13 assume-unchanged + 없음 = 삭제 | `AH6_assume_unchanged_keeps_the_index_blob` | `test_assume_unchanged_does_not_hide_a_deletion` |
| blob id 가 git 의 것 | `AH16_the_digest_is_not_a_blob_id` | `test_an_untouched_worktree_is_an_empty_comparison` (첫 판 **생존** — 판정은 git 이 내용을 견주어 그대로였다) |
| B14 다른 것만 쓴다 | `AH17_every_blob_is_written` | 같은 시험 (첫 판 **생존**) |
| loose 객체 형식 | `AH18_the_loose_header_is_wrong` | 23 시험(첫 셋: clean 필터 · `=` 드라이버 · 공백 드라이버) |
| B15·B16 `:`·`"` 든 객체 디렉터리 | `AH25_an_alternate_with_a_colon_is_taken` | `test_an_object_directory_git_cannot_take_as_an_alternate_is_named` |
| 실제 저장소를 대체 저장소로 | `AH15_the_real_store_is_not_an_alternate` | 91 시험 |
| `core.splitIndex=false` | `AH8_the_split_index_leaks` | `test_the_snapshot_writes_nothing_into_the_repository_and_runs_nothing_it_configures` |
| `core.hooksPath=/dev/null` | `AH9_the_hooks_run` | 같은 시험 |
| `SNAPSHOT_PINS`(fsmonitor) | `AH10_the_fsmonitor_runs` | 같은 시험 |
| `ls-files` 에도 고정 | `AH20_the_ls_files_runs_the_fsmonitor` | 같은 시험 |
| B17 `update-index` 실패 | `AH26_a_failed_update_index_is_ignored` | `test_every_git_fault_of_the_snapshot_is_named` |

**B5·B6**(`-z` 끝의 빈 조각 건너뛰기)은 변이를 넣지 않았다 — 빼면 빈 조각이 B7·B8 의 결함이 되어 **모든** 판정이 멈추므로
모든 실저장소 시험이 잡는다(따로 잴 값이 없다).
