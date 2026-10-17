# Branch Test Map: `_measure_landing_inputs` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 목록을 한 번 읽는다 | W1 | `test_the_landing_inputs_read_each_bundle_once` |
| 하한이 못 서면 수리 신호 `None` | W19 | `test_repairs_that_were_never_measured_are_not_read_as_none` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 증거를 스스로 안 읽는다 | W1 | `test_one_command_reads_the_evidence_once_to_judge 외 5` |
| 하한이 못 서면 수리 신호 `None` | W19 | `test_repairs_that_were_never_measured_are_not_read_as_none` |
