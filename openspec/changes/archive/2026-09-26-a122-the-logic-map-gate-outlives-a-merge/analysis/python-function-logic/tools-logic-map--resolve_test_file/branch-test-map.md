# Branch Test Map: `resolve_test_file` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 고르기 세 갈래가 깔때기를 지난다 | AA4 | `test_a_change_that_becomes_open_while_judged_asks_for_a_rerun`(같은 깔때기) · 기존: `test_a_qualified_path_that_does_not_exist_is_not_silently_skipped` |

## task 7.5.9 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 이름 붙은 경로 | (기존) | `test_a_qualified_path_resolves_to_that_file_not_the_local_one` |
| B2 · B3 이름 붙은 경로도 추적 파일만 | AM2 | `test_a_qualified_coordinate_into_an_untracked_file_is_unresolved` |
| B4 · B5 패키지 안도 추적 파일만 | AM3 | `test_an_untracked_file_in_the_package_does_not_answer_for_the_tracked_one` |
| B6 맨이름은 추적 목록에서 | AM4(트리 `rglob`) | 3 — `test_an_untracked_copy_does_not_hide_a_bare_name_coordinate` · 패키지 시험 · 구조 `test_no_read_primitive_lives_outside_the_funnels` |
| B7 모호하면 해소 안 함 | (기존) | — (정책 · 사본 시험의 대조군이 한 개일 때의 거절을 본다) |

## 보수 (독립 리뷰 둘, 2026-09-26)

짝은 변이로 잡았다(`75_mut.py` 창 `246:253` · `233:234`, 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 416`; 사본은 최종 파일과 주석 한 덩이만 다르고 AST 가 같다 — 실측).

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 새 B2 · B3 경로를 풀어서 대조(`..` · 심링크) | MR1(`root / cited` 그대로) | 2 — `test_a_dot_dot_or_doubled_slash_coordinate_names_the_tracked_file` · `test_a_coordinate_through_a_tracked_directory_link_names_the_tracked_file` |
| 새 B3 저장소 밖이면 `None` | MR6(`ValueError` 를 올림) | `test_a_coordinate_that_leaves_the_repository_is_unresolved` |
| 추적 목록 멤버십(B4 이하 — 7.5.9 표) | AM6 재조준(`ls-tree HEAD`) | 16 |

## task 7.5.12 (2026-09-27)

짝은 변이로 잡았다(`75_mut.py` 창 `267:276` · 재조준 `87:88` · `240:242` · `247:248`, 사본 ext4 · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 454`).

| 갈래 | 변이 | 잡는 시험(수) |
|---|---|---|
| B2 의 풀이는 깔때기로 | MW9 | 1 — 구조 **하나뿐** |
| B2 경로를 푼다 | MR1(재조준) | 2 — `..` · 심링크 디렉터리 |
