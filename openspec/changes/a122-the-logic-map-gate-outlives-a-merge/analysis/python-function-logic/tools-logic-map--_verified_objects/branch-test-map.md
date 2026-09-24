# Branch Test Map: `_verified_objects` (Python, a122 task 7.5.34 — 새 함수)

변이 로그(`75_mut.py`, 창 `96:143 · 143:190`, 대조군 양끝 GREEN `Ran 372` — 하네스 사본은 `docs/WORKFLOW.md` 가 없어 skip 2)에서 옮겼다. 변이마다 값 하나. 이 함수는 변이 판 뒤에 바뀌지 않았다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 물을 것이 없으면 git 을 안 부른다 | `AJ1_nothing_asked_still_runs_git` | `test_every_bad_answer_of_the_object_store_is_named` · `test_every_git_fault_of_the_comparison_is_named` |
| B3 `cat-file` 실패 | `AJ2_a_failed_cat_file_is_read` | `test_every_bad_answer_of_the_object_store_is_named` · `test_every_git_fault_of_the_comparison_is_named` |
| B5 머리 줄 없음 | `AJ3_a_missing_header_is_read` | `test_every_bad_answer_of_the_object_store_is_named` |
| B6·B7 다른 oid | `AJ4_the_answer_may_name_another_oid` | `test_every_bad_answer_of_the_object_store_is_named` |
| B6·B7 다른 종류 | `AJ5_the_answer_may_be_another_kind` | `test_every_bad_answer_of_the_object_store_is_named` |
| B6·B7 크기가 숫자 아님 | `AJ6_the_size_is_not_read_as_a_number` | `test_every_bad_answer_of_the_object_store_is_named` |
| B8 몸이 모자람 | `AJ7_a_short_body_is_read` | `test_every_bad_answer_of_the_object_store_is_named` |
| B9 끝 표시 | `AJ8_the_terminator_is_not_checked` | `test_every_bad_answer_of_the_object_store_is_named` |
| B10 **해시 검증** | `AJ9_the_hash_is_not_checked` | 4 시험: `test_a_forged_base_object_is_refused` · `test_a_sparse_entry_that_differs_from_the_base_is_judged_from_its_verified_blob` · `test_every_bad_answer_of_the_object_store_is_named` 외 1 |
| B10 해시에 머리(`<종류> <크기>\0`)를 넣는다 | `AJ10_the_hash_omits_the_header` | 179 시험(모든 검증이 틀린 oid 를 계산해 거절한다) |
| B11 남은 바이트 | `AJ11_leftover_bytes_are_ignored` | `test_every_bad_answer_of_the_object_store_is_named` |

B2 · B4(요청 조립 · 대답 순회)는 따로 잴 값이 없다 — 빠지면 모든 검증이 멈춘다.
