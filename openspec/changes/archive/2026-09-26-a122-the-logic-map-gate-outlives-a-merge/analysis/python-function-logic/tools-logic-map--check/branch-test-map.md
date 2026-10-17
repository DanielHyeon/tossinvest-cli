# Branch Test Map: `check` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 착지가 판정한 바이트에 묶기 | W7 | `test_evidence_swapped_after_the_landing_is_judged_is_refused` |
| 대상에 한 번 읽은 바이트 | W8 | `test_evidence_swapped_after_the_verdict_read_it_is_not_read_again` |
| 열거형 호출 판정도 그 바이트 | W10 | `test_every_evidence_read_goes_through_the_one_reader` |
| 이관 change 미리 읽기 선별 격리 | W12 | `test_an_adopted_changes_escaping_bundle_keeps_its_name` |
| 조언 결함 격리 | W13 | `test_advice_that_cannot_be_computed_does_not_replace_the_verdict` |
| 앞선 실행의 사실 지우기 | W20 | `test_a_reused_context_does_not_carry_the_last_verdicts_evidence` |
| 착지 뒤 읽기 결함 = 판정 줄 | — | `test_a_fault_reading_the_landing_for_the_verdict_is_a_verdict` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 조기 반환은 한 번 읽은 값 | Y13 | `test_evidence_deleted_after_the_landing_is_judged_is_not_read_as_no_evidence` |
| 대상 목록은 한 번 읽은 값 | Y14 | `같은 시험` |
| 판정은 다시 안 읽는다 | Y18 | `test_one_command_reads_the_evidence_once_to_judge 외 3` |
| 대상 판정에 같은 바이트 | W8 | `test_evidence_swapped_after_the_landing_is_judged_is_not_what_the_verdict_reads 외 5` |
| 열거형 호출에 같은 바이트 | W10 | `test_every_evidence_read_goes_through_the_one_reader 외 1` |
| 목록 못 열면 이름 댄 줄 | Y17 | `test_a_bundle_directory_that_can_be_entered_but_not_listed_is_named` |
| 앞선 실행의 사실 전부 지움 | W20 | `test_a_reused_context_carries_nothing_from_the_last_run` |
| 이관 미리 읽기 선별 격리 | W12 | `test_an_adopted_changes_escaping_bundle_keeps_its_name` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B1 앞 — 사실을 **맨 앞**에서 비운다(`facts.clear()`) | Z10 | `test_a_context_is_emptied_even_when_the_id_does_not_resolve` |
| B26 한 출구의 재확인 — 달라졌으면 판정 대신 "다시 돌려라" | Z5 | `test_evidence_rewritten_after_the_read_is_not_reported_as_passing` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 원장을 열고 닫는다 | Z5 | `test_evidence_rewritten_after_the_read_is_not_reported_as_passing` 외 14 |
| 판정 앞의 거절은 대조하지 않는다 | — | `test_a_context_is_emptied_even_when_the_id_does_not_resolve`(오타 id 는 그 문장으로 멈춘다) |
