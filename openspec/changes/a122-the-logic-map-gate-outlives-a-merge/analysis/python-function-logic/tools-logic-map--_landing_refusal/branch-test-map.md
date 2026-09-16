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

## task 7.2.2 — 고정 소스를 바꾼 착지 (변이로 채운 행만, `722_mut.py` 양성 대조군 GREEN)

| 갈래 (소스 한 줄) | 덮는 테스트 | 덮이나 |
|---|---|---|
| `if all(_committed_bytes(root, base, source) == _committed_bytes(root, candidate, source) for source in sources)` → 거절 | `test_the_recorder_refuses_evidence_written_before_the_edit` · `test_a_declared_landing_that_changes_none_of_its_pinned_sources_is_refused` · `test_the_recorder_refuses_side_branch_evidence_merged_after_the_work` · `test_a_change_whose_work_precedes_its_base_gets_no_landing` · `test_it_does_not_record_the_base_even_when_the_base_matches` | yes — K-A(가드 삭제) 다섯 모두 빨갛다 |
| 같은 줄의 `all` (하나 이상 바뀌면 받는다) | `test_one_changed_pinned_source_is_enough` | yes — K-B(`any`)는 **이 시험 하나**만 잡는다. 첫 판 픽스처는 읽기 전용 파일을 base 뒤에 만들어 도달하지 못했고 K-B 가 살아남았다 — 도달 단언을 넣은 뒤 잡힌다 |
| `sources = sorted({source for _, source, _ in _pinning_bundles(root, analysis)})` (전 번들) | `test_one_changed_pinned_source_is_enough` | yes — K-G(`[:1]`) 같은 시험 하나 |
| 비교 기준 `base` | 25 시험 | yes — K-C(`HEAD`) |
| 위치: 미보유 뒤 | `test_evidence_written_at_the_base_cannot_declare_the_base` 외 넷 | yes — K-D(하한 앞으로) |

## task 7.2.6 — 착지 뒤의 자기 수리 (변이로 채운 행만, `726_mut.py` 무변이 대조군 GREEN)

| 갈래 (소스 한 줄) | 덮는 테스트 | 덮이나 |
|---|---|---|
| `later = [commit for commit in repairs if not _is_ancestor(root, commit, candidate)]` → 거절 | `test_a_review_fix_to_a_pinned_file_after_the_record_is_refused` · `test_any_go_file_counts_not_only_the_pinned_one` · `test_the_signal_survives_archiving_the_change` 외 4 | yes — S1(가드 삭제) 일곱이 빨갛다 |
| 후보 자신은 안 센다 (`_is_ancestor` 는 같은 커밋에서 참) | `test_the_repair_commit_itself_is_the_landing_when_it_refreshes_the_evidence` | yes — S5. 첫 판 시험은 불일치 가드에 먼저 걸려 **닿지 않았고**(픽스처가 번들을 안 갱신), 수리와 갱신을 한 커밋에 넣은 모양으로 바꾼 뒤 잡힌다 |
| 이름은 **가장 오래된** 수리 | `test_the_oldest_repair_is_the_one_named` | yes — S7 |
| 위치: K2 뒤, 수락 앞 | `test_a_candidate_its_evidence_does_not_describe_keeps_that_sentence` | yes — S6(맨 앞으로). 첫 판에는 이 시험이 없어 SURVIVED 였다 |
| 호출자가 **한 번** 재서 넘긴다 (두 집이 아니다) | `test_the_recorder_will_not_write_a_landing_its_own_later_work_outruns`(계산) · `test_the_refusal_says_how_to_move_the_record`(선언) | yes — S10 · S11 각각 |
