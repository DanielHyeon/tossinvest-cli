# Branch Test Map: `_verdict` (Python, a122 task 7.5.2.2)

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B9 목록 · 파일을 못 읽으면 이름 댄 줄(조용히 건너뛰지 않는다) | Y17 | `test_a_bundle_directory_that_can_be_entered_but_not_listed_is_named` |
| B10 이름은 `filename` 이 **글자일 때만** — 아니면 번들 이름 | Z13 · Z17 | `test_a_read_failure_without_a_path_names_the_bundle` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 목록 실패는 이름 댄 줄 | Y17 · AA12 | `test_a_bundle_directory_that_can_be_entered_but_not_listed_is_named` 외 6 |
| 파일 실패는 파일 이름으로 | Z13 | `test_a_fifo_among_the_bundle_files_is_named_not_skipped` 외 5 |
| 이름이 글자가 아니면 번들 이름 | Z17 | `test_a_read_failure_without_a_path_names_the_bundle` |
| 열거형 호출 판정도 같은 바이트 | W10 | `test_every_evidence_read_goes_through_the_one_reader` 외 2 |
