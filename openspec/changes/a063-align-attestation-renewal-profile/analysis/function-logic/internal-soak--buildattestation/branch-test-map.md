# Branch Test Map: `BuildAttestation`

| B-id | Scenario | Focused test and bounded claim |
|---|---|---|
| B1 | Incomplete summary returns `IncompleteError`. | `TestBuildAttestationRefusesAnIncompleteSoak`. |
| B2 | Iteration over read-only successful endpoints. | `TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise` exercises it; zero endpoints unmeasured. |
| B3 | Reject a non-GET endpoint. | `TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise`. |
| B4 | Propagate `acceptSupervised` error. | unmeasured by focused test. |
| B5 | Append accepted supervised endpoints. | unmeasured by focused test. |
| B6 | Prefix a nonblank note. | unmeasured by focused test. |
