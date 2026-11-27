# Branch Test Map: `reasonText`

Source: `internal/operatorview/exit_line.go` (198-248) — AST branches 24. RED/GREEN 은 a207 시험의 실제 실행(`analysis/tdd/`)으로만 yes 로 적고, 기존 시험은 이 change 에서
선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch strings.TrimSpace(reason) {` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B2 | `case "engine_not_running":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B3 | `case "snapshot_quarantined":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B4 | `case "observation_older_than_limit":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B5 | `case "observation_in_future":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B6 | `case "invalid_observed_at":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B7 | `case "not_evaluated_yet":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B8 | `case "no_saved_evaluation":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B9 | `case "legacy_snapshot_absent", "legacy_event":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B10 | `case "repeat_of_refused_proposal":` | `TestA207TheRepeatReasonReadsAsASummarisedRepeatedRefusal` | yes | yes |
| B11 | `case "invalid_stored_snapshot", "invalid_effective_snapshot":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B12 | `case "legacy_policy_identity_unknown", "legacy_adoption_context_required":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B13 | `case "partial_snapshot_tuple", "partial_evaluated_tuple", "partial_seed_tuple", "partial_tuple":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B14 | `case "partial_policy_tuple", "invalid_policy_identity":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B15 | `case "flattened_snapshot_mismatch":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B16 | `case "ambiguous_exit_evidence":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B17 | `case "exit_evidence_unlinked":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B18 | `case "lineage_cycle":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B19 | `case "lineage_ambiguous":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B20 | `case "lineage_depth_exceeded":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B21 | `case "lineage_scope_mismatch":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B22 | `case "invalid_event_evidence":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B23 | `case "":` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
| B24 | `default:` | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | no | no |
