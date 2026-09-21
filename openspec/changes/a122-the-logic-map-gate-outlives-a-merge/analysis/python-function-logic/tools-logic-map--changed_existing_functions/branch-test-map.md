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
