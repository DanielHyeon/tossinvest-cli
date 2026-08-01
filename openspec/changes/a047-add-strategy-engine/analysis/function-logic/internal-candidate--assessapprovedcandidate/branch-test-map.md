# Branch Test Map: `AssessApprovedCandidate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid threshold set | `TestAssessApprovedCandidateFailsClosed/invalid_set` | compile RED in a047 review lane | yes |
| B2 | invalid key/first sighting | `TestAssessApprovedCandidateFailsClosed/{invalid_candidate_life,zero_first_seen}` | existing | yes |
| B3 | cooled, expired, stale, exact expiry, active re-cross | `TestAssessApprovedCandidateFailsClosed` current-life rows | yes | yes |
| B4 | threshold scope mismatch | `TestAssessApprovedCandidateFailsClosed/wrong_market` | existing | yes |
| B5 | one or more vetoes are raised | `TestAssessApprovedCandidateFailsClosed/dangerous` | existing | pass |
| B6 | one or more veto inputs are unmeasured | `TestAssessApprovedCandidateFailsClosed/unmeasured` | existing | pass |
| B7 | aggregate chase is not a full pass despite no raised/unmeasured list | fail-closed table/invariant coverage | existing | pass |
| Scenario | all checks pass and immutable current-life provenance is sealed | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | yes | pass |
