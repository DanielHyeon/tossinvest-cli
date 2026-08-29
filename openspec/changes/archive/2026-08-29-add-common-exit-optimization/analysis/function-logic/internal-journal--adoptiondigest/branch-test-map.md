# Branch Test Map: `adoptionDigest`

- Source: `internal/journal/adoption.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` path at line 458 and its complement/boundary | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests | yes | yes |
