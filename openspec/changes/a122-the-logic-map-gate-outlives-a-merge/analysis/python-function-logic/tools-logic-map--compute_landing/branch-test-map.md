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
