# Branch Test Map: `_evidence_fingerprint` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 바이트가 지문에 | W2 | `test_a_bundle_rewritten_during_the_walk_is_seen 외 30` |
| `HEAD` 가 지문에 | W3 | `test_a_commit_landing_mid_walk_is_seen_when_recording · …_when_judging` |
| 고정 목록이 지문에(미리 읽기가 쓴다) | W4 | `test_the_cost_does_not_grow_with_the_number_of_pinned_sources` |
