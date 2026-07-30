# Branch Test Map: `Journal.OpenExitStates`

- Source: `internal/journal/apply_hook.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 596 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
| B2 | `for` path at line 602 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
| B3 | `if` path at line 604 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
| B4 | `if` path at line 609 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
