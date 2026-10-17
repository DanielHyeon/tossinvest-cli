# Branch Test Map: `_bundle_text` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 52 변이 · 생존 0, 첫 판 생존 W23 은 **안 닿음** → 닿는 시험을 더해 잡음). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `ast.json` 은 한 번 읽은 것만 | W11 | `test_every_evidence_read_goes_through_the_one_reader` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(`75_mut.py` 76 변이 · 생존 0 · 무변이 대조군 초록). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| `ast.json` 은 목록과 무관 | W11 | `test_a_bundle_listing_that_misses_ast_json_keeps_the_bytes_judged` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B2 `ast_raw is not None` — 목록과 무관하게 판정한 바이트 | W11 | `test_a_bundle_listing_that_misses_ast_json_keeps_the_bytes_judged` |
| B6 `FileNotFoundError` 만 건너뛴다 — 못 읽는 정규 파일은 올린다 | Z2 | `test_an_unreadable_bundle_file_is_named_not_skipped` |
| B7 `raw is None` → 건너뛴다(FIFO · 장치) — 행동으로는 안 보인다(`io.BytesIO(None)` 이 빈 버퍼라 Z14 가 **도달한 채 살아남았다**), 구조로 못 박았다 | Z14 | `test_the_bundle_reader_skips_a_non_regular_file_before_it_decodes` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 못 읽는 파일은 이름 댄 줄 | Z2 | `test_an_unreadable_bundle_file_is_named_not_skipped` 외 6 |
| 못 푸는 바이트도 이름 댄 줄 | Z14 | `test_a_bundle_file_that_is_not_utf8_is_named_not_skipped` · `test_the_bundle_reader_has_one_silent_skip_and_it_is_for_what_is_gone` |
| 한 번 읽은 `ast.json` 바이트를 쓴다 | W11 | `test_a_bundle_listing_that_misses_ast_json_keeps_the_bytes_judged` · `test_the_bundle_text_is_built_from_the_bytes_the_command_read` |
| 건너뛰는 모양은 '사라진 파일' 하나 | — | 구조: `test_the_bundle_reader_has_one_silent_skip_and_it_is_for_what_is_gone` |
