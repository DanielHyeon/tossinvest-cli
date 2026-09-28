# 반전 대상 시험 (편집 전 AST 로 셈)

모집단: `tools/logic-map/test_check_analysis.py` 의 함수 중 본문에 `(?i)adopt|execution_baseline|a063|audited` 가 있는 것. `test_execution_baseline.py`(모듈 단위 시험 전체)는 D1 에 따라 삭제한다.

| 클래스.함수 | 줄 | 참조 수 | 처분 |
|---|---|---:|---|
| CheckAnalysisTests.test_main_distinguishes_ordinary_adoption_and_invalid_results | 154-175 | 13 | TBD |
| CheckAnalysisTests.test_main_real_adoption_prints_exception_label_only_after_validation | 177-195 | 10 | TBD |
| CheckAnalysisTests._adoption_with_complete_bundle | 413-453 | 7 | TBD |
| CheckAnalysisTests.test_real_adoption_without_local_venv_accepts_external_doctor_probe | 456-472 | 12 | TBD |
| CheckAnalysisTests.test_valid_adoption_uses_e_and_requires_complete_current_bundle | 474-481 | 7 | TBD |
| CheckAnalysisTests.test_real_adoption_deleted_function_requires_base_revision | 483-490 | 7 | TBD |
| CheckAnalysisTests.test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back | 492-504 | 9 | TBD |
| CheckAnalysisTests.test_real_adoption_stale_current_ast_hash_fails | 506-512 | 6 | TBD |
| CheckAnalysisTests.test_real_adoption_rejects_local_maps_with_function_logic_reference | 514-521 | 6 | TBD |
| CheckAnalysisTests._adoption_output | 523-530 | 2 | TBD |
| CheckAnalysisTests.test_the_adoption_window_ends_at_the_audited_source_commit | 532-555 | 9 | TBD |
| CheckAnalysisTests.test_the_adoption_path_does_not_advise_a_record_it_would_refuse | 557-570 | 6 | TBD |
| CheckAnalysisTests.test_a_landing_record_is_refused_in_the_adoption_path | 572-591 | 9 | TBD |
| CheckAnalysisTests.test_the_recorder_refuses_the_adoption_path_too | 593-609 | 8 | TBD |
| CheckAnalysisTests._archive_adoption | 611-621 | 5 | TBD |
| CheckAnalysisTests.test_an_archived_adoption_is_rechecked_by_its_id | 623-643 | 12 | TBD |
| CheckAnalysisTests.test_a_copied_adoption_record_does_not_make_another_change_a063 | 645-666 | 11 | TBD |
| CheckAnalysisTests.test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised | 668-688 | 10 | TBD |
| ADeclaredLandingMustBePinnedByEvidence.test_an_absolute_bundle_path_still_pins_a_landing | 3498-3532 | 1 | TBD |
| AnEmptyRequiredSetIsAnnouncedNotSwallowed.test_empty_required_with_bundles_is_reported_and_still_passes | 4161-4200 | 2 | TBD |
| TheVerdictReadsWhatTheLandingJudged.test_an_adopted_changes_escaping_bundle_keeps_its_name | 5771-5787 | 8 | TBD |
| OneCommandJudgesOneHistoryAndOneRead.test_a_reused_context_carries_nothing_from_the_last_run | 6550-6567 | 7 | TBD |
| EveryGateSubprocessHasATimeout.test_every_child_process_in_the_gate_modules_names_a_timeout | 9911-9923 | 1 | TBD |
| TheSuiteDoesNotReadTheDevelopersGitConfig.test_the_execution_baseline_suite_pins_itself | 10408-10420 | 3 | TBD |
