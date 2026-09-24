# Branch Test Map: `changed_existing_functions` (Python, a122 task 7.5.2.4)

## task 7.5.2.4 — 통합 diff 의 본문 줄은 파일 헤더가 아니다 (7.5.2.3 재리뷰, 2026-09-21 보안)

짝은 손으로 고르지 않았다 — 각 갈래를 깨는 변이를 `75_mut.py` 에 넣고 **실제로 빨개진 시험**을 옮겼다
(`AB1~AB5` 전부 CAUGHT, 총 118 · 앵커 전수 정확히 1회 · 창 앞뒤 무변이 대조군 GREEN · `Ran 304` 동일).
픽스처는 지어낸 diff 문자열이 아니라 **진짜 저장소 · 진짜 `git diff`** 다 — 첫 줄의 시험이 그것부터 못 박는다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 문법 사실(본문에 헤더 모양 줄이 실제로 나온다) | — | `test_git_really_emits_a_file_header_shape_from_an_ordinary_source_line` |
| B31 본문에서는 이름을 안 읽는다 | `AB1_body_lines_name_the_file_again` | `test_a_deleted_header_shaped_line_does_not_erase_the_required_function` · `test_an_added_header_shaped_line_does_not_downgrade_the_requirement` · `test_a_deleted_sql_comment_does_not_become_the_name_of_a_base_file` |
| B32 첫 훅이 본문을 연다 | `AB2_the_body_never_opens` | 같은 셋 |
| B30 새 파일이 본문을 닫는다 | `AB3_a_new_file_does_not_close_the_body` | `test_a_real_deleted_file_header_still_means_there_is_no_current_logic` |
| B31 본문의 둘째 훅도 센다 | `AB4_only_the_first_hunk_of_a_file_counts` | 결함 시험 셋 |
| B34 `/dev/null` 은 이름이 아니다 | `AB5_dev_null_is_an_ordinary_name` | `test_a_real_added_file_header_still_means_there_is_no_base_logic` |
| B36 지워진 파일에 현재 지도를 안 요구한다 | — (B34 와 대칭, 같은 시험이 값을 단언한다) | `test_a_real_deleted_file_header_still_means_there_is_no_current_logic` |

**안 건드린 갈래**: `B1~B25`(훅 적재 앞의 거절 · `flush()` 의 base/현재 통과)는 이 로트가 안 바꿨다 —
기존 시험 21개가 그대로 서 있고, `AB3` 이 그중 둘을 빨갛게 만든 것이 그 사실의 덤 증거다.

## task 7.5.25 — 스냅숏 (2026-09-24)

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 껍데기 B2 워킹트리만 스냅숏 | `AH19_a_commit_target_takes_the_snapshot` | `test_a_commit_target_does_not_read_the_worktree`(원장 단언; 첫 판 **생존**·"못 쟀다" — 판정은 같고 원장만 달랐다) — 7.5.31 뒤 6 시험(커밋 대상이 스냅숏을 열면 git 밖 대조가 커밋 대상에서 돈다) |
| `_compared` 워킹트리 = `--cached` | `AH1_the_worktree_is_seen_through_git` | 9 시험(7.5.31 창) — 문 시험들(clean · sparse · …) |
| 판정 diff 가 스냅숏 환경을 받는다 | `AH3_the_judgement_skips_the_snapshot` | 25 시험(7.5.31 창) |
| 몸통 B13~B15 현재 쪽 = 스냅숏 바이트 | `AH4_the_current_side_rereads_the_disk` | `test_the_current_side_is_judged_from_the_snapshot_bytes` |

지운 갈래(편집 전 B43~B45 · raise 6·7)의 변이 `AF7` · `AF11` · `AG3~AG5` 는 하네스에서 뺐다 — 겨눌 코드가 없다.

## task 7.5.34

편집 전 B2(`if target:` — 커밋 대상은 스냅숏 없이)와 그 변이 AH19 는 없어졌다. 커밋 대상이 워킹트리를 안 읽는 것은
`_isolated_comparison` B11 이 지키고 `test_a_commit_target_does_not_read_the_worktree`(원장 단언)가 본다. 남은 B1(base 필수)은
7.5.25 의 행 그대로다.

### 몸통 `_changed_existing_functions` — task 7.5.34 (창 `174:187` 재실행 · `143:190`)

| 갈래(편집 후 몸통) | 변이 | 잡는 시험 |
|---|---|---|
| 판정 diff 가 격리한 환경에서 돈다 | `AJ49_the_judgement_runs_outside_the_comparison` | 148 시험 |
| B5 옛 쪽 바이트가 비교에 없다 → 이름 댄 결함 | `AJ50_the_old_side_is_not_checked` | `test_a_deleted_go_gitlink_is_refused_by_name` · `test_base_file_load_failure_is_not_treated_as_new_file` |
| B13·B14 현재 쪽은 비교의 바이트 | `AJ51_the_current_side_is_empty` | 50 시험 |
| 훅의 옛 쪽 · 새 쪽 좌표 | `AC5_the_hunk_reads_the_base_side_twice` | `test_the_hunk_keeps_the_new_side_coordinates_apart_from_the_base_side` (창 `143:190`) |

편집 전 B40(git 밖 대조)과 그 변이 AI6 은 없어졌다. 편집 전 AH3 · AH4 가 재던 자리는 AJ49 · AJ51 이 새 코드에서 잰다.
