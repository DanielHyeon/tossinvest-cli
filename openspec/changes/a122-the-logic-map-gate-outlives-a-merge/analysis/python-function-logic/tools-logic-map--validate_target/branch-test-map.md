# Branch Test Map: `validate_target` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `evidence` 를 받으면 다시 안 읽는다 | W8 | `test_evidence_swapped_after_the_verdict_read_it_is_not_read_again · test_a_bundle_that_appears_after_the_verdict_read_is_missing` |
| 못 읽음 = 판정 줄 | W23 | `test_evidence_that_could_not_be_read_is_named_not_skipped` |
| 모양 검사 | W17 | `test_a_malformed_ast_is_a_verdict_not_a_traceback` |
| 미리 읽기 fallback 은 착지에서 | — | `test_a_source_missing_from_the_prefetch_is_read_at_the_landing` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 못 읽음 = 판정 줄 | W23 | `test_evidence_that_could_not_be_read_is_named_not_skipped` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B7 산문도 `_read_regular` 로 | Z4 | `test_a_fifo_in_place_of_a_required_file_is_named_not_waited_on` |
| B9 `OSError` → `could not be read`(`missing` 아님) | Z15 | `test_an_unreadable_required_file_is_not_called_missing` |
| B10 `raw is None` → `could not be read` | Z4 | `test_a_fifo_in_place_of_a_required_file_is_named_not_waited_on` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 산문도 종류를 보고 연다 | Z4 | `test_a_fifo_in_place_of_a_required_file_is_named_not_waited_on` |
| 못 읽음 ≠ 없음 | Z15 | `test_an_unreadable_required_file_is_not_called_missing` |
| 못 푸는 산문은 대상 이름 댄 줄 | AA11 | `test_prose_that_is_not_utf8_names_the_target` |
| 워킹트리 소스도 깔때기로 · 고리는 '못 읽는다' | Y28 | `test_a_symlink_loop_in_a_bundle_source_is_named` · `test_a_pinned_source_rewritten_after_it_was_judged_asks_for_a_rerun` |
