# Branch Test Map: `_landing_refusal`

번호가 아니라 **소스 한 줄**로 적는다 — [[positional-branch-ids-break-hand-renumbering]].
짝은 손으로 고르지 않았다: 각 갈래를 지우는 변이(`76_mut.py` R1~R6, R16)를 걸어 **실제로 빨개진**
시험을 옮겼다. 하네스는 이름을 셋까지만 찍는다 — 넷 이상이면 "외 N".

| 갈래 (소스) | 변이 | 선언 경로에서 잡는 시험 | 계산 경로에서 잡는 시험 |
|---|---|---|---|
| `if not _is_ancestor(root, base, candidate)` | R1 | `test_a_landing_before_the_base` | `test_a_commit_the_base_never_reached_is_not_the_landing` · `test_what_it_records_on_a_merged_history_the_gate_still_accepts` |
| `if not pinning` | R2 | `test_a_zero_bundle_change_cannot_declare_a_landing` · `test_base_revision_bundles_do_not_pin_a_landing` · `test_a_window_that_could_not_be_derived_is_not_printed` | **닿지 않는다(구조)** — `compute_landing` 은 번들 0 이면 순회 **전에** 자기 사유로 돌아간다 |
| `if mismatched` | R3 | `test_a_landing_the_evidence_does_not_describe` · `test_borrowed_evidence_refuses_a_landing_it_does_not_describe` (외 1) | `test_the_landing_is_the_first_matching_commit_not_where_evidence_entered` |
| `if not floor` | R4 | `test_uncommitted_evidence_cannot_pin_a_landing` | **닿지 않는다(구조)** — 하한이 없으면 순회 **전에** 돌아간다. 그 조기 종료 문장은 7.8 의 M7 시험이 잰다 |
| `if not _is_ancestor(root, floor, candidate)` | R5 · R16 | `test_a_stale_landing_is_refused_on_a_log_follow_machine` · `test_evidence_written_at_the_base_cannot_declare_the_base` | `test_a_candidate_older_than_its_evidence_is_not_the_landing` — **7.6 이 추가**. 첫 판에서 R16(계산 경로가 하한 대신 base 를 넘김)이 SURVIVED 였다 |
| `if unheld` | R6 | `test_a_bundle_symlinked_out_of_the_watched_path_cannot_pin_a_landing` · `test_a_placeholder_committed_early_cannot_pin_a_landing` | `test_the_gate_will_not_compute_a_landing_it_cannot_hold` |
| `return ("", [])` (받음) | R7 · R9 | 모든 양성 대조군 — 예: `test_the_computed_value_is_still_accepted` | `test_the_recorded_value_is_one_the_gate_then_accepts` 외 왕복 시험 전부 |

**"닿지 않는다(구조)" 둘은 결함이 아니다.** 규칙 안의 그 두 조건은 선언 경로를 위해 있고,
계산 경로는 같은 조건을 순회 전에 한 번 보고 자기 문장으로 돌아간다. 둘을 규칙에서 지우면
선언 경로의 시험이 빨개지므로(R2·R4 CAUGHT) 규칙은 여전히 못 박혀 있다.
