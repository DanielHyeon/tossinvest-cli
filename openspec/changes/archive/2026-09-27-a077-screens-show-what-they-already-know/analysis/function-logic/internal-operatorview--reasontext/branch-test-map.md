# Branch Test Map: `reasonText`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the switch over the trimmed reason maps to its own sentence | every case below | n/a | pass |
| B2 | `engine_not_running` (a077) maps to its own sentence | `TestAStoppedEngineClosesTheProtectionLine` | yes | yes |
| B3 | `snapshot_quarantined` (a077) maps to its own sentence | `TestAQuarantinedPositionIsNotDrawnAsProtected` | yes | yes |
| B4 | `observation_older_than_limit` maps to its own sentence | `TestAConsoleWithNoEngineMarkerKeepsTheObservationAgeBound` | n/a | pass |
| B5 | `observation_in_future` maps to its own sentence | `TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted` | n/a | pass |
| B6 | `invalid_observed_at` maps to its own sentence | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | n/a | pass |
| B7 | `not_evaluated_yet` maps to its own sentence | existing SEED fixture test | n/a | pass |
| B8 | `no_saved_evaluation` maps to its own sentence | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | n/a | pass |
| B9 | `legacy_snapshot_absent`, `legacy_event` maps to its own sentence | `TestPositionsRenderCanonicalExitLineFixtures` | n/a | pass |
| B10 | `invalid_stored_snapshot`, `invalid_effective_snapshot` maps to its own sentence | `TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted` | n/a | pass |
| B11 | `legacy_policy_identity_unknown`, `legacy_adoption_context_required` maps to its own sentence | existing legacy tests | n/a | pass |
| B12 | `partial_snapshot_tuple` and the other partial tuples maps to its own sentence | existing corruption tests | n/a | pass |
| B13 | `partial_policy_tuple`, `invalid_policy_identity` maps to its own sentence | existing corruption tests | n/a | pass |
| B14 | `flattened_snapshot_mismatch` maps to its own sentence | existing corruption tests | n/a | pass |
| B15 | `ambiguous_exit_evidence` maps to its own sentence | existing orders evidence tests | n/a | pass |
| B16 | `exit_evidence_unlinked` maps to its own sentence | existing orders evidence tests | n/a | pass |
| B17 | `lineage_cycle` maps to its own sentence | existing lineage tests | n/a | pass |
| B18 | `lineage_ambiguous` maps to its own sentence | existing lineage tests | n/a | pass |
| B19 | `lineage_depth_exceeded` maps to its own sentence | existing lineage tests | n/a | pass |
| B20 | `lineage_scope_mismatch` maps to its own sentence | existing lineage tests | n/a | pass |
| B21 | `invalid_event_evidence` maps to its own sentence | existing orders evidence tests | n/a | pass |
| B22 | the empty reason maps to its own sentence | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | n/a | pass |
| B23 | the default arm maps to its own sentence | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` | n/a | pass |

`TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` in `internal/operatorview/exit_line_test.go`
enumerates the reason codes and is the guard that a new one cannot be added
without a sentence.
