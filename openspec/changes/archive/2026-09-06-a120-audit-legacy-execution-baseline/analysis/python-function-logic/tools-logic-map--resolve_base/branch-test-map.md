# Branch Test Map: `resolve_base`

Pre-edit coverage map. Test names describe the existing baseline contract; new negative fixtures remain implementation work and are not claimed as completed.

| Branch | Scenario | Baseline test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Read failure for `base-commit.txt` | ``CheckAnalysisTests.test_invalid_base_fails_closed`` | not run in Stage 1B | not run in Stage 1B |
| B2 | `git rev-parse --verify` fails for either persisted candidate or override | ``CheckAnalysisTests.test_environment_base_cannot_override_persisted_change_base`` | not run in Stage 1B | not run in Stage 1B |
| B3 | An environment override resolves differently from persisted base | ``CheckAnalysisTests.test_environment_base_cannot_override_persisted_change_base`` | not run in Stage 1B | not run in Stage 1B |
