# Branch Test Map: `AssessApprovedCandidate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid threshold set | `TestAssessApprovedCandidateFailsClosed/invalid_set` | compile RED in a047 review lane | yes |
| B2 | invalid key/first sighting | `TestAssessApprovedCandidateFailsClosed/{invalid_candidate_life,zero_first_seen}` | existing | yes |
| B3 | cooled, expired, stale, exact expiry, active re-cross | `TestAssessApprovedCandidateFailsClosed` current-life rows | yes | yes |
| B4 | threshold scope mismatch | `TestAssessApprovedCandidateFailsClosed/wrong_market` | existing | yes |
| B5 | veto raised or unmeasured | `TestAssessApprovedCandidateFailsClosed/{dangerous,unmeasured}` | existing | yes |
| B6 | pass preserves state/last-seen/exclusive validity | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | yes | yes |
| B7 | final measured-and-clear invariant | same immutable provenance pass/refusal table | existing | yes |
