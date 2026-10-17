# Branch Test Map: `section_name` (Python, 7.5.13 — 안쪽 함수)

## task 7.5.13 (2026-09-27)

짝은 변이로 잡았다(`75_mut.py` 창 `253:262` · `117:118` · 재실행 `257:258` · `261:262`, 사본 `A122_HARNESS_WORK` ext4 · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 427` / 재실행 `Ran 429`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 · B2 구역이 레코드보다 많으면 결함 | MQ5 — **첫 판 생존**(~~정상 git 은 수를 맞춘다~~ 정정: typechange 는 레코드 하나 · 구역 둘 — 진짜 git 시험을 더했다). 시험을 더해 재실행 CAUGHT | `test_a_section_beyond_the_listed_files_is_named` |
| B3 머리 줄과 렌더링 대조 | MQ1(`if False`) | `test_the_two_views_are_compared_name_by_name` |

## 마감 수리 (적대 리뷰, 2026-09-27)

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| B1 · B2 (진짜 git) | MQ5 재실행 | `test_a_typechange_has_more_sections_than_records_and_is_named` — typechange 만(레코드 1 · 구역 2) |
| B3 (진짜 git) | MQ1 재실행 | `test_a_typechange_before_another_file_is_named_by_the_header` — typechange 뒤 `y.go` (구역 2 가 `y.go` 레코드와 대조) |
| B3 의 접두사 절반 | **MX1**(리뷰어 변이 — 접두사 무시, 429 초록 생존) | `test_the_prefix_is_part_of_the_comparison`(판정 diff 의 `--src-prefix` 를 `c/` 로 바꿔 치움) |
