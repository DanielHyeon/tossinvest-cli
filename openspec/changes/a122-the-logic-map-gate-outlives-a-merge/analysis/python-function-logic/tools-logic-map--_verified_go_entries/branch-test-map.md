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
| B17(7.5.35 뒤 B25) 하위 트리를 걷는다 | `AJ18_subtrees_are_not_walked` | 16 시험(첫 셋: `test_a_failure_prints_the_comparison_window` · `test_a_fault_while_asking_the_recorder_does_not_become_advice` · `test_a_forged_base_object_is_refused`) |
| B18(7.5.35 뒤 B26) `.go` 만 | `AJ19_the_base_keeps_every_file` | 90 시험 |

B1 · B3 · B4 · B8 · B11 · B14(순회)는 따로 잴 값이 없다 — 빠지면 base 가 비어 모든 판정이 바뀐다.

## task 7.5.35 (창 `157:209` · 모든 창 `0:53` · `53:105` · `105:157` 대조군 양끝 GREEN `Ran 376`, 하네스 사본은 skip 2)

분기 번호는 **편집 후**(`ast.after-7.5.35.json`)다. 위 표의 B17 · B18 은 편집 전 번호이고 지금은 B25 · B26 이다 — 같은 변이
AJ18 · AJ19 가 다시 잡혔다(이 절의 판 — 재리뷰 전 코드 · 376 시험 — 에서 19 · 91, 최종 판 `Ran 377` 에서 20 · 91).

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B18 모드 다섯 밖 → 거절 | `AK1_any_mode_is_walked` | `test_a_tree_git_would_read_differently_is_refused_by_name` · `test_a_commit_target_tree_git_would_read_differently_is_refused_by_name` |
| 모드 표의 `100755` | `AK2_the_executable_mode_is_refused` | 118 시험 |
| 모드 표의 `120000` | `AK3_a_symlink_is_refused` | `test_every_tree_git_writes_is_accepted` |
| 모드 표의 `160000` | `AK4_a_gitlink_is_refused` | `test_every_tree_git_writes_is_accepted` · 하위 모듈 시험 둘 |
| B19 · B20 이름 → 거절 | `AK5_names_are_not_checked` | `test_a_tree_git_would_read_differently_is_refused_by_name` |
| B19 빈 이름 | `AK6_an_empty_name_passes` | `test_a_tree_git_would_read_differently_is_refused_by_name`(빈 이름의 트리 — 첫 판 SURVIVED, 아래) |
| B19 `/` 든 이름 | `AK7_a_slash_in_a_name_passes` | `test_a_tree_git_would_read_differently_is_refused_by_name` |
| B19 `.` | `AK8_dot_passes` | 같은 시험 |
| B19 `..` | `AK9_dot_dot_passes` | 같은 시험 |
| B21 같은 이름 → 거절 | `AK10_duplicates_pass` | 같은 시험(파일과 트리가 한 이름) |
| B22 순서 → 거절 | `AK11_order_is_not_checked` | 같은 시험 |
| B17 트리 열쇠의 `/` | `AK12_order_ignores_the_tree_slash` | `test_every_tree_git_writes_is_accepted`(`a.go` 가 디렉터리 `a` 앞) |
| 열쇠를 넘긴다(`last = key`) | `AK13_the_order_does_not_advance` | `test_a_tree_git_would_read_differently_is_refused_by_name` |
| 이름을 기억한다(`names.add`) | `AK14_names_are_not_remembered` | 같은 시험 |
| B23 사유가 있으면 멈춘다 | `AK15_the_refusal_is_silent` | 트리 거절 시험 둘 |

**AK6 이 첫 판에서 살아남았다.** 시험은 "canonical form" 만 봤고, 빈 이름의 **파일**은 열쇠 `b""` 가 앞의 것보다 클 수 없어
순서 검사(B22)가 같은 문장으로 대신 거절했다. 빈 이름의 **트리**는 열쇠가 `b"/"` 라 순서 검사를 지나므로 이름 검사만
잡는다. 그래서 시험에 그 모양("empty tree name")을 더하고 모양마다 **사유**까지 단언하게 바꿨다 — 이 수정은 병행 세션
(tossos-d6)의 팀메이트가 이 로트를 고아로 오판해 12:17 에 먼저 적었고, 내용을 확인해 그대로 받았다. 최종 시험으로 AK6 을
다시 돌렸다(재리뷰 전 376 시험 파일 — 최종 377 시험의 증거는 아래 창 `159:212`): 창 `192:193` **CAUGHT**(`test_a_tree_git_would_read_differently_is_refused_by_name`, 대조군 양끝 GREEN `Ran 376`). `key <= last` → `<` 는 변이에 넣지 않았다: 같은 이름은 B21 이 먼저 잡아 같은 열쇠가 B22 에 닿지 않는다.

### 7.5.35 적대 재리뷰 뒤 (최종 코드 · 최종 시험, 212 변이)

재리뷰가 셋을 더 셌고 코드와 시험을 고쳤다 — `100664` 를 받고(P2-1), 붙어 있지 않은 같은 이름을 시험에 더하고(P2-2), gitlink
정렬 이웃을 정상 트리 시험에 더했다(P3-1). 그 셋을 변이로 못 박았다:

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 모드 표의 `100664`(초기 git) | `AK21_early_git_mode_is_refused` | `test_a_tree_from_early_git_is_accepted` |
| B21 본 이름 **전부**와 견준다(옆만이 아니라) | `AK22_duplicates_only_next_to_each_other`(리뷰어의 M11) | `test_a_tree_git_would_read_differently_is_refused_by_name`(붙어 있지 않은 같은 이름) |
| B17 gitlink 는 파일처럼 견준다 | `AK23_a_gitlink_sorts_like_a_tree`(리뷰어의 M1) | `test_every_tree_git_writes_is_accepted`(gitlink `b` · `b-.go`) |

전체 212 를 최종 코드에서 다시 돌렸다(창 `0:53` · `53:106` · `106:159` · `159:212` 각 53/53, 대조군 양끝 GREEN `Ran 377`). 그
판은 루트 파일시스템이 가득 찬 시간(병행 세션 통보, 약 13:30~15:30 KST)과 겹쳤다 — 시험 수가 첫 판보다 **더 많이** 실패한
24 변이 중 1 넘게 늘어난 15(AA15 · AA16 · AA17 · AA19 · AJ26–AJ29 · T6 · T7 · U1 · Y17–Y20 — 정확히 +1 인 9 는 새 시험 하나로 설명된다)와 새 AK21–AK23 을 디스크가 돌아온 뒤 따로 다시
돌렸다: 18/18 CAUGHT, 실패 수는 첫 판과 같다(U1 21 → 1 등 — 부풀린 것은 디스크 오류였다).
