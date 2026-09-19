# Branch Test Map: `_read_evidence` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 못 읽음 → `None` | W23 | `test_evidence_that_could_not_be_read_is_named_not_skipped` |
| 모든 읽기가 여기를 거친다 | W8 · W9 · W10 · W11 | `test_every_evidence_read_goes_through_the_one_reader` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 번들은 `iterdir` | Y15 | `test_a_bundle_directory_that_can_be_entered_but_not_listed_is_named 외 3` |
| 없음 ≠ 못 읽음 | Y16 | `test_absent_and_unreadable_evidence_are_different_reads · test_a_bundle_that_appears_after_the_verdict_read_is_missing` |
