# Branch Test Map: `resolve_base` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 없는 것과 못 읽는 것을 가른다 | AA15 | `test_a_fifo_in_place_of_the_base_commit_is_named_not_waited_on` · `test_the_base_commit_rewritten_after_it_was_judged_asks_for_a_rerun` |

## task 6.4(b) — 창의 시작도 끝처럼 잠근다 (2026-09-26)

짝은 변이 로그에서 옮겼다(`75_mut.py 212:214` · `214:220`, 대조군 창 양끝 GREEN `Ran 384` · 생존 0). 로그는
`analysis/harness/_work/mut64.*.log`(커밋 안 함) — 아래 이름은 그 로그의 실패 시험 이름 그대로다.

| 갈래 (`ast.after-6.4.json`) | 변이 | 잡는 시험 |
|---|---|---|
| B4 40자리 모양 | AL2 | `test_a_name_or_a_short_id_is_not_a_base` |
| B5·B6 커밋된 값과 다른 디스크 | AL3(검사 삭제) | `test_an_uncommitted_edit_cannot_move_a_committed_base` · `test_a_borrowed_base_is_frozen_too` |
| B5 의 `committed is not None` (HEAD 에 없는 새 base 는 받는다) | AL4 | `test_an_uncommitted_new_base_is_read_from_disk` 외 10 |
| `_committed_bytes` 가 명령의 `head` 를 쓴다 | AL6(`"HEAD"` 로) | `test_every_history_read_uses_the_one_resolved_commit` · `test_no_history_read_names_symbolic_head_outside_the_pin` (구조 시험 — 행동은 같다) |
| B9 태그 객체 id | AL5 | `test_the_id_of_a_tag_object_is_not_a_commit_id` |

## task 6.4 보수 — 이동이 대조를 벗지 못한다 (2026-09-26)

번호는 `ast.after-64r.json`(옛 `ast.before-64r.json` 과 `difflib` 로 정렬 — 옛 B7~B16 → 새 B9~B18, 옛 B5·B6 은 새 B5~B8 로 대체).
짝은 변이 로그에서 옮겼다: `75_mut.py 212:228`(사본 `A122_HARNESS_WORK` = ext4 스크래치, 무변이 대조군 창 양끝 GREEN `Ran 388` ·
**16/16 CAUGHT** · 생존 0, 18 판 18분 50초). 위 6.4(b) 표의 `B5·B6` 행은 이 표로 대체된다(번호가 바뀌었다).

| 갈래 (`ast.after-64r.json`) | 변이 | 잡는 시험 |
|---|---|---|
| B4 40자리 모양 — `fullmatch` | AL2(삭제) · AL9(`match`) | `test_a_name_or_a_short_id_is_not_a_base`(AL9 는 새 `<sha>~1` 행이 잡는다 — 첫 판에서 **살아남았다**, P2 X1) |
| B5 else — 같은 id 의 다른 자리 | AL10(`else []`) | `test_a_moved_directory_keeps_its_committed_base` |
| B5 의 `committed is not None`(어디에도 없으면 받는다) | AL4(빈 `held` 를 불일치로) | `test_an_uncommitted_new_base_is_read_from_disk` · `test_another_id_elsewhere_in_head_is_not_consulted` 외 8 |
| B6 · B7 대조 | AL3(`if True: continue`) | 편집 · 빌린 base · 이동 셋 |
| B7 이동의 불일치 | AL11(이동이면 받는다) | `test_a_moved_directory_keeps_its_committed_base` |
| B8 제자리 문장 · 이동 문장 | AL15(`if False`) | `test_an_uncommitted_edit_cannot_move_a_committed_base` · `test_a_borrowed_base_is_frozen_too` |
| `_committed_elsewhere` B1 ls-tree 결함 | AL16(`if False`) | `test_a_failed_listing_is_a_fault_not_an_empty_archive` |
| `_committed_elsewhere` 아카이브 열거 | AL12(`if False`) | `test_a_moved_directory_keeps_its_committed_base`(A3 · 아카이브 → 활성) |
| `_committed_elsewhere` 활성 자리 | AL13(목록에서 뺌) | `test_a_moved_directory_keeps_its_committed_base`(A1 · A2) |
| `_committed_elsewhere` id 전체 일치 | AL14(`endswith`) | `test_another_id_elsewhere_in_head_is_not_consulted` |
| `_committed_bytes` 가 명령의 `head` | AL6 | 구조 시험 둘(행동은 같다) |
| B11 태그 객체 id | AL5 | `test_the_id_of_a_tag_object_is_not_a_commit_id` |
