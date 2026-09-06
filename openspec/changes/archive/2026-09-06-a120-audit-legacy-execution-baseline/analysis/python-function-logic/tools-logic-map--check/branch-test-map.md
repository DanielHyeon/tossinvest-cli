# Branch Test Map: `check`

Pre-edit coverage map. Test names describe the existing baseline contract; new negative fixtures remain implementation work and are not claimed as completed.

| Branch | Scenario | Baseline test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Base/function derivation raises | ``CheckAnalysisTests.test_explicit_exemption_is_accepted`` | not run in Stage 1B | not run in Stage 1B |
| B2 | Function-logic reference exists | ``no focused current test; adoption/reference fixture required`` | not run in Stage 1B | not run in Stage 1B |
| B3 | No local/reference analysis directory | ``CheckAnalysisTests.test_explicit_exemption_is_accepted`` | not run in Stage 1B | not run in Stage 1B |
| B4 | Analysis directory is empty | ``no focused current test; existing `check` negative-path fixture required`` | not run in Stage 1B | not run in Stage 1B |
| B5 | Each bundle target | ``no focused current test; existing `check` negative-path fixture required`` | not run in Stage 1B | not run in Stage 1B |
| B6 | Each required changed-existing function | ``CheckAnalysisTests.test_modified_function_cannot_use_exemption`` | not run in Stage 1B | not run in Stage 1B |
