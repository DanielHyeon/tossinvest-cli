# Branch Test Map: `changed_existing_functions`

Pre-edit coverage map. Test names describe the existing baseline contract; new negative fixtures remain implementation work and are not claimed as completed.

| Branch | Scenario | Baseline test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | No explicit base | ``no focused current test; new immutable-target fixture required`` | not run in Stage 1B | not run in Stage 1B |
| B2 | `git diff` fails | ``CheckAnalysisTests.test_git_diff_failure_is_not_treated_as_empty_change`` | not run in Stage 1B | not run in Stage 1B |
| B3 | No file/hunks pending in nested `flush` | ``CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing`` | not run in Stage 1B | not run in Stage 1B |
| B4 | Base file cannot be loaded | ``CheckAnalysisTests.test_base_file_load_failure_is_not_treated_as_new_file`` | not run in Stage 1B | not run in Stage 1B |
| B5 | Current path exists | ``CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing`` | not run in Stage 1B | not run in Stage 1B |
| B6 | Current function was absent at base | ``CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing`` | not run in Stage 1B | not run in Stage 1B |
| B7 | Current function intersects a new-side hunk | ``CheckAnalysisTests.test_modified_function_cannot_use_exemption`` | not run in Stage 1B | not run in Stage 1B |
| B8 | Diff parser sees file header, old/new source, or hunk | ``CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing`` | not run in Stage 1B | not run in Stage 1B |
