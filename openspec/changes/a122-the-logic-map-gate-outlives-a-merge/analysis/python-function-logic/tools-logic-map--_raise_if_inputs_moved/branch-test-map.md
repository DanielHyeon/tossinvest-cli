# Branch Test Map: `_raise_if_inputs_moved` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 수락 직전 | V9 | `test_a_bundle_rewritten_during_the_walk_is_seen 외 6` |
| 선언 경로 거절 앞 | W5 | `test_a_declared_landing_refused_on_moving_evidence_says_run_again` |
