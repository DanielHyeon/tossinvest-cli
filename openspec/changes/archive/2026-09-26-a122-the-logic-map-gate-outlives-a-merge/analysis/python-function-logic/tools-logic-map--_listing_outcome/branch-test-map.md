# Branch Test Map: `_listing_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 **113 · 생존 0**: 112 는 한 판에서 CAUGHT(스위트 297 · 무변이 대조군 GREEN) · AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(298)). 이 로트는 하네스도 고쳤다: 사본을 **프로세스별**로 가르고(배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼은 일이 있었다 — 무변이 대조군이 빨개져서야 알았다), 사본이 원본과 같은지 단언하고, 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심게 했다(새 줄의 머리를 찾던 판본은 한 줄 통째 교체에서 언제나 '안 닿음' 이라고 답했다). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| 지문에 이름과 **종류**가 든다 | AA19 | `test_a_name_that_turns_into_a_bundle_while_judged_asks_for_a_rerun` (이 시험이 없을 때 **살아남았다**) |

## task 7.5.10 · 7.5.11 (2026-09-26)

짝은 손으로 고르지 않았다 — 갈래마다 변이를 걸어 실제로 빨개진 시험을 옮겼다(`analysis/harness/75_mut.py` 창 `228:246` · `97:98` · `99:100` · `101:102` · `102:103` · `112:113` · `239:240` · `244:246`(재실행), `7515_mut.py`; 사본 `A122_HARNESS_WORK` · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 407`, 구조 시험을 더한 뒤 창 `102:103` · `244:246` 은 `Ran 408`). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 | 잡는 시험 |
|---|---|---|
| 지문이 단사 — `\n` · `\t` 이음이 아니다 (7.5.10) | AN1 | `test_the_known_collisions_have_different_fingerprints` · `test_a_listing_that_becomes_its_collision_while_judged_asks_for_a_rerun` |
| 길이 접두 (7.5.10) | AN2 | `test_the_known_collisions_have_different_fingerprints`(셋째 쌍 `{a, b}` 대 `{afb}`) |
| 종류 한 글자 (B7) | AA19(재조준) | `test_a_name_that_turns_into_a_bundle_while_judged_asks_for_a_rerun` |
| B2~B4 종류를 `os.stat` 으로 (7.5.11) | AO1(`is_dir()`) | 4 — 끊긴 링크 둘 · 고리 · 구조 `test_the_kind_is_not_read_through_is_dir` |
| B4 못 물으면 결함 | AO2("파일" 로 삼킴) | 3 — 끊긴 링크(번들 사이 · 번들 안) · 고리 |
| B4 결함의 타입이 `UnstatableEntry` | AO3(원래 `OSError` 를 그대로 — 번들 사이면 면제 경로로 샌다) | 3 — 같은 셋 |
| 따라가서 디렉터리면 디렉터리 | — | `test_a_link_to_a_bundle_is_still_a_bundle`(양성 대조, 편집 전에도 초록) |

## task 7.5.38 (2026-09-27)

짝은 변이로 잡았다(`75_mut.py` 창 `267:276` · 재조준 `87:88` · `240:242` · `247:248`, 사본 ext4 · pid 별 · 무변이 대조군 창 양끝 GREEN `Ran 454`).

| 갈래 | 변이 | 잡는 시험(수) |
|---|---|---|
| 새 B5 · B6 목록 뒤 사라진 항목은 움직임 | MW5(`if False`) | 3 — 목록 수준 · 증거 디렉터리 판정 · 구조 |
| 새 B6 의 `lexists` 가 끊긴 링크를 가른다 | MW6(ENOENT 면 전부 "사라짐") | 5 — 끊긴 링크 시험들 · 양성 대조 |
| B2~B4 `os.stat` 으로 묻는다 | AO1 · AO2(재조준) | 9 · 5 |
