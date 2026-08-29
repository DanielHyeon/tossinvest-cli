# Branch Test Map: `scanExitState`

- Source: `internal/journal/apply_hook.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 553 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
| B2 | `if` path at line 556 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
| B3 | `if` path at line 560 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
| B4 | `if` path at line 563 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
