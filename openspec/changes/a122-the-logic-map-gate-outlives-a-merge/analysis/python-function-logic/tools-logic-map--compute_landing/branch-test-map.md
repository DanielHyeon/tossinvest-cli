# Branch Test Map: `compute_landing` · `record_landing`

번호가 아니라 **소스 한 줄**로 적는다 — [[positional-branch-ids-break-hand-renumbering]].

| 함수 | 갈래 | 덮는 테스트 | 덮이나 |
|---|---|---|---|
| `compute_landing` | 번들 0 → 사유 | `test_it_refuses_when_no_evidence_pins_a_landing` | yes |
| `compute_landing` | 맞는 커밋 없음 → 사유 | — 픽스처 없음(실물 a089·a095 로만 실측) | **no** |
| `compute_landing` | 가장 낮은 후보가 base 여도 안 쓴다 | `test_it_does_not_record_the_base_even_when_the_base_matches` | yes |
| `compute_landing` | 바닥 없음 → 사유 | — 픽스처 없음 | **no** |
| `record_landing` | 기록이 이미 있다 | `test_it_refuses_to_overwrite_an_existing_record` | yes |
| `record_landing` | 워킹트리가 더럽다 | `test_it_refuses_while_tracked_files_are_modified` | yes |
| `record_landing` | 빌린 증거 | — 픽스처 없음 | **no** |
| `record_landing` | 이관 change | — 픽스처 없음(실물 a063 로만 실측) | **no** |
| `record_landing` | 성공 → 왕복 | `test_the_recorded_value_is_one_the_gate_then_accepts` | yes |

**덮이지 않는 넷을 적는 이유는 그것이 이 편집의 결론이 아니기 때문이다.** 이 표는
커버리지를 주장하지 않고 열거한다 — [[passing-test-is-not-evidence]].

## task 6.5 — 걷기 실패 사유 (변이로 채운 행만)

위 표(6.1.2 시점)의 `맞는 커밋 없음 → 사유 — 픽스처 없음 · no` 는 **낡았다** — 7.8 이 위조 픽스처를 기록 경로에 줬고
(`test_the_recorder_refuses_the_forgery_this_class_is_named_after`), 6.5 가 두 후보 픽스처를 더했다. 아래 행은 변이가 실제로
빨갛게 만든 시험만 적는다(`65_mut.py`, 양성 대조군 GREEN).

| 함수 | 갈래 (소스 한 줄) | 덮는 테스트 | 덮이나 |
|---|---|---|---|
| `compute_landing` | `return ('', f'no commit at or after the evidence ({floor[:12]}) matches every pinning bundle — at the first commit walked, {first}')` | `test_the_recorder_names_what_already_differs_where_the_evidence_entered` · `test_the_recorder_refuses_the_forgery_this_class_is_named_after` | yes — N1(옛 꼬리) · N3(인용 제거) · N4(이름 있을 때만 인용) 셋 다 **두 시험 모두** 빨갛다 |
| `compute_landing` | `first = first or refusal` (첫 후보만 남긴다) | `test_the_recorder_names_what_already_differs_where_the_evidence_entered` | yes — N2(마지막 후보 인용)는 **이 시험 하나**만 잡는다. 위조 픽스처는 후보가 하나라 첫/마지막을 못 가른다 |
