# a147 implementation verification

Implementation authorization follows the earlier design-only review. All executions below use synthetic files and an isolated checkout based on `936d7e0d`; no new market session or external request was started.

- Preregistration, required fields, timestamp, immutable output and manifest digest: `test_manifest_limits`, `test_required_plan_fields_and_horizon`, `test_exclusive_write`, `test_append_no_overwrite_and_symlink`, `test_fixture_end_to_end_and_missing_capture`.
- MFI14/Δ3, minimum 18, 0/100/undefined, epsilon, invalid OHLCV, future/duplicate/nonconsecutive/stale and offset: `test_rising_and_boundary`, `test_all_negative_flow_is_zero`, `test_down_and_change`, `test_invalid_and_temporal`, `test_epsilon_and_invalid_open`, `test_zero_flow_and_stale`, `test_offset_required`.
- Source identity, source/timezone, as-of and session boundary: `test_capture_identity_and_asof`, `test_fixture_end_to_end_and_missing_capture`. Source and receipt authenticity still require independent evidence.
- Label gap, missing, late, early, invalid and exact 60-minute timing: `test_label_due_and_delay`.
- Pairing/unknown selection, candidate conflict, slot normalization, equal-cycle average, incomplete and interrupted sessions: `test_unknown_and_inconsistent_remain_in_denominator`, `test_same_slot_and_symbol_collision_is_unknown`, `test_skew_normalizes_without_joining_next_cycle`, `test_equal_cycle_weight_and_unknown_counts`, `test_fixture_end_to_end_and_missing_capture`.
- Output/input separation, exclusive creation, no-follow and hard-link rejection: `test_prevent_symlink_output`, `test_exclusive_write`, `test_append_no_overwrite_and_symlink`. The CLI also refuses output under the repository.

`python3 -m unittest discover -s tools/a147-jev-mfi-offline -p 'test*.py' -v`: 19 tests, exit 0, `/tmp/a147-primitives.log`. Synthetic actual CLI output `/tmp/a147-cli-evidence/report-frozen.json`; `/tmp/a147-cli-{record,evaluate}-frozen.exit` are 0. Test subprocesses override socket connect to fail; this is not an OS-level network-isolation proof. The complete aggregate gate has not yet run; its final exit is recorded separately.

This is a new Python leaf; no existing function is edited, and the existing Go AST/branch-map tool is not applicable. Core CodeGraph freshness was verified before edit in this isolated checkout; a final refresh must follow all document changes. Advisory GBrain is separate from hard freshness.
