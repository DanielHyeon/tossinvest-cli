# Branch Test Map: `_verdict` (Python, a122 task 7.5.2.2)

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B9 목록 · 파일을 못 읽으면 이름 댄 줄(조용히 건너뛰지 않는다) | Y17 | `test_a_bundle_directory_that_can_be_entered_but_not_listed_is_named` |
| B10 이름은 `filename` 이 **글자일 때만** — 아니면 번들 이름 | Z13 · Z17 | `test_a_read_failure_without_a_path_names_the_bundle` |
