# Branch Test Map: `validate_target` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `evidence` 를 받으면 다시 안 읽는다 | W8 | `test_evidence_swapped_after_the_verdict_read_it_is_not_read_again · test_a_bundle_that_appears_after_the_verdict_read_is_missing` |
| 못 읽음 = 판정 줄 | W23 | `test_evidence_that_could_not_be_read_is_named_not_skipped` |
| 모양 검사 | W17 | `test_a_malformed_ast_is_a_verdict_not_a_traceback` |
| 미리 읽기 fallback 은 착지에서 | — | `test_a_source_missing_from_the_prefetch_is_read_at_the_landing` |
