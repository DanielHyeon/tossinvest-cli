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
