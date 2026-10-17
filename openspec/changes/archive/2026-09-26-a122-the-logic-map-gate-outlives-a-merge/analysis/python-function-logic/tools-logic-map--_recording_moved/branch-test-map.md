# Branch Test Map: `_recording_moved` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 역사를 **먼저** 묻는다 | AA8 | `test_a_commit_landing_mid_record_writes_nothing` |
| 거절 집합 전체를 다시 묻는다 | AA7 | `test_no_record_is_written_when_the_tree_turns_dirty_during_the_walk` |
| 그리고 원장을 견준다 | Z8 | `test_the_evidence_is_read_again_only_to_accept` 외 2 |

## task 7.5.16 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 역사가 먼저 | AA8 — **첫 판 생존**(이 로트의 B2 수리가 `_judged_state_moved` 의 역사 물음을 거절 앞에 세워 출력이 같아졌다). 구조 시험을 더해 재실행 CAUGHT | `test_the_history_is_asked_before_the_refusal_is_computed` |
| B2 원장이 거절보다 먼저 | AQ1(편집 전 순서) | 2 — `test_an_input_that_moved_during_the_walk_is_named_before_the_refusal_it_causes` · 구조 시험 |
| B2 원장을 묻는다 | AQ2 | 3 — 위 둘 · `test_a_bundle_whose_read_fails_once_is_never_judged_absent` |
| 거절도 여전히 계산하고 말한다 | AA7(재조준) | 3 — `test_a_refusal_with_no_moved_input_is_still_said` · `test_no_record_is_written_when_the_tree_turns_dirty_during_the_walk` 외 |
