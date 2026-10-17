# 반전 대상 시험 (편집 전 AST 로 셈)

모집단: `tools/logic-map/test_check_analysis.py` 의 함수 중 본문에 `(?i)adopt|execution_baseline|a063|audited` 가 있는 것. `test_execution_baseline.py`(모듈 단위 시험 전체)는 D1 에 따라 삭제한다.

| 클래스.함수 | 줄 | 참조 수 | 처분 |
|---|---|---:|---|
| CheckAnalysisTests.test_main_distinguishes_ordinary_adoption_and_invalid_results | 154-175 | 13 | 반전 → `test_main_prints_only_the_ordinary_success_line` |
| CheckAnalysisTests.test_main_real_adoption_prints_exception_label_only_after_validation | 177-195 | 10 | 반전 → `test_a063_with_a_leftover_record_is_judged_from_its_base_commit`(출력에 이관 라벨 없음) |
| CheckAnalysisTests._adoption_with_complete_bundle | 413-453 | 7 | 대체 → 모듈 없는 `_a063_fixture`(정적 기록 · 리터럴 id) |
| CheckAnalysisTests.test_real_adoption_without_local_venv_accepts_external_doctor_probe | 456-472 | 12 | 삭제 — 지운 `adoption.validate` 를 부르는 skip 시험. `SDD_PYTHON` 은 `tools/sdd/test_sdd_doctor.py` 가 덮는다(freeze F3) |
| CheckAnalysisTests.test_valid_adoption_uses_e_and_requires_complete_current_bundle | 474-481 | 7 | 반전 → `…judged_from_its_base_commit`(기준 P) · `test_a063s_modified_function_is_still_required` |
| CheckAnalysisTests.test_real_adoption_deleted_function_requires_base_revision | 483-490 | 7 | 반전 → `test_a_deleted_function_needs_a_base_revision_bundle_from_the_base_commit` |
| CheckAnalysisTests.test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back | 492-504 | 9 | 반전 → `test_sdd_base_ref_accepts_only_the_base_commit`(P 수락 · E 거절 · 깨진 기록 무시) |
| CheckAnalysisTests.test_real_adoption_stale_current_ast_hash_fails | 506-512 | 6 | 반전 → `test_a_stale_current_hash_fails` |
| CheckAnalysisTests.test_real_adoption_rejects_local_maps_with_function_logic_reference | 514-521 | 6 | 반전 → `test_a_reference_beside_local_maps_is_refused` |
| CheckAnalysisTests._adoption_output | 523-530 | 2 | 대체 → `_a063_cli` |
| CheckAnalysisTests.test_the_adoption_window_ends_at_the_audited_source_commit | 532-555 | 9 | 반전 → `…judged_from_its_base_commit`(landing "" · 워킹트리 창) |
| CheckAnalysisTests.test_the_adoption_path_does_not_advise_a_record_it_would_refuse | 557-570 | 6 | 반전 → `test_the_advice_offers_a063_the_landing_command`(codex C3) |
| CheckAnalysisTests.test_a_landing_record_is_refused_in_the_adoption_path | 572-591 | 9 | 반전 → `test_a063_records_and_is_judged_by_the_landing_rules` · `test_a063s_landing_record_is_held_to_the_computed_value` |
| CheckAnalysisTests.test_the_recorder_refuses_the_adoption_path_too | 593-609 | 8 | 반전 → `test_a063_records_and_is_judged_by_the_landing_rules`(기록 rc 0) · `test_own_go_work_after_a063s_landing_is_refused` |
| CheckAnalysisTests._archive_adoption | 611-621 | 5 | 흡수 → `test_an_archived_a063_is_rechecked_by_its_id_on_the_general_path` 안의 `git mv` |
| CheckAnalysisTests.test_an_archived_adoption_is_rechecked_by_its_id | 623-643 | 12 | 반전 → `test_an_archived_a063_is_rechecked_by_its_id_on_the_general_path` |
| CheckAnalysisTests.test_a_copied_adoption_record_does_not_make_another_change_a063 | 645-666 | 11 | 반전 → `test_a_copied_record_changes_nothing_for_another_id` |
| CheckAnalysisTests.test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised | 668-688 | 10 | 반전 → `test_an_undecodable_landing_record_is_named_by_the_landing_rules` |
| ADeclaredLandingMustBePinnedByEvidence.test_an_absolute_bundle_path_still_pins_a_landing | 3498-3532 | 1 | 무변경 — 주석 한 줄의 a063 언급 |
| AnEmptyRequiredSetIsAnnouncedNotSwallowed.test_empty_required_with_bundles_is_reported_and_still_passes | 4161-4200 | 2 | 무변경 — `audited source-commit` 부재 단언은 그대로 참 |
| TheVerdictReadsWhatTheLandingJudged.test_an_adopted_changes_escaping_bundle_keeps_its_name | 5771-5787 | 8 | 수정 → `_a063_fixture` 로 일반 경로에서 같은 사실 |
| OneCommandJudgesOneHistoryAndOneRead.test_a_reused_context_carries_nothing_from_the_last_run | 6550-6567 | 7 | 수정 → 옛 키가 **없음**을 단언(`assertNotIn`) |
| EveryGateSubprocessHasATimeout.test_every_child_process_in_the_gate_modules_names_a_timeout | 9911-9923 | 1 | 수정 → 모듈 목록 `(check_analysis,)` + 디렉터리의 자식 프로세스 모듈 정확 집합 단언(freeze F9) |
| TheSuiteDoesNotReadTheDevelopersGitConfig.test_the_execution_baseline_suite_pins_itself | 10408-10420 | 3 | 삭제(대상 모듈 삭제) → `test_the_exception_module_and_its_suite_are_gone` 로 흡수 |

추가(반전이 아닌 새 핀): `test_a_leftover_record_changes_neither_verdict_nor_context_nor_output`(D2 · 1.1 논증 제거 뒤 핀) · `test_an_irregular_leftover_record_is_not_read_either`(freeze F9).
