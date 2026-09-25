# Branch Test Map: `_isolated_comparison` (Python, a122 task 7.5.34 — 새 함수)

변이 로그(`75_mut.py`, **최종 코드**에서 다시 돈 창 `174:187` · `135:136`, 대조군 양끝 GREEN `Ran 372`)에서 옮겼다. 첫 판(창 `143:190`)에서 이 함수의 두 변이가 살아남았다 — AJ44(출처를 고르는 갈래, 같은 oid 면 같은 바이트라 고를 것이 없었다 → 한 표로 바꿨다) · AJ47(두 diff 의 `GIT_INDEX_FILE`, 트리 둘을 견주는 diff 는 인덱스를 안 쓴다 → 줄을 지웠다). 그 뒤 이 함수를 겨누는 변이를 전부 다시 돌렸다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B3 옵션 모양의 리비전 | `AJ38_an_option_is_taken_as_a_revision` | `test_every_git_fault_of_the_comparison_is_named` |
| B6 `rev-parse` 실패 | `AH21_a_failed_rev_parse_is_read` | `test_a_forged_root_tree_or_commit_is_refused` · `test_every_git_fault_of_the_comparison_is_named` (창 `135:136` 재실행) |
| B7·B8 `rev-parse` 답 모양 | `AJ39_a_short_rev_parse_is_read` | `test_every_git_fault_of_the_comparison_is_named` |
| B12 다른 경로만 | `AJ40_unchanged_paths_are_loaded` | `test_an_untouched_worktree_is_an_empty_comparison` |
| B14 새 쪽(커밋 대상 · sparse)도 검증해 읽는다 | `AJ41_the_target_side_is_not_fetched` | 77 시험 |
| B16·B17 워킹트리 쪽은 저장소에 안 묻는다 | `AJ42_the_worktree_side_is_fetched_from_the_store` | 33 시험(첫 셋: clean 필터 · `=` 드라이버 · 공백 드라이버) |
| B16 gitlink 는 안 묻는다 | `AJ43_a_gitlink_is_fetched` | `test_a_deleted_go_gitlink_is_refused_by_name` |
| B19 디스크 바이트를 표에 싣는다 | `AJ44_the_disk_bytes_are_not_tabled` | 99 시험 |
| B24 물려받은 대체 저장소를 지운다 | `AJ45_an_inherited_alternate_is_kept` | `test_an_inherited_alternate_does_not_reach_the_comparison` |
| 격리한 저장소(`GIT_OBJECT_DIRECTORY`) | `AJ46_the_store_is_the_real_one` | 33 시험 |

B1 · B2 · B4 · B5 · B9 · B10 · B13 · B15 · B20–B23(리비전 조립 · 순회 · 문장 이름)은 따로 잴 값이 없다 — 빠지면 모든 판정이 바뀌거나 문장만 바뀐다. B22·B23 의 gitlink 건너뛰기는 AJ43 의 짝이다(물으면 `has no blob`, 쓰면 `KeyError`).

## task 7.5.35 (창 `157:209`, 대조군 양끝 GREEN `Ran 376`)

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B24 물려받은 대체 저장소를 지운다(앵커를 새 조건에 맞춰 옮김) | `AJ45_an_inherited_alternate_is_kept` | `test_an_inherited_alternate_does_not_reach_the_comparison` |
| B24 `GIT_DIFF_OPTS` 를 지운다 | `AK20_git_diff_opts_is_passed_on` | `test_the_judged_diff_format_does_not_follow_the_user_config` |

나머지 AJ38–AJ46 은 다시 잡혔다(같은 시험; AJ42 · AJ46 은 7.5.34 의 33 → 이 절의 판(재리뷰 전 코드 · `Ran 376`) 35 → 최종 판(`Ran 377`) 36 시험).
