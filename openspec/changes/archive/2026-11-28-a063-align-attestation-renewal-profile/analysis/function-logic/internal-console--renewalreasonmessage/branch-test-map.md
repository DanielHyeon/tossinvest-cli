# Branch Test Map: `renewalReasonMessage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | dispatch | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability` | n/a | yes |
| B2 | record_empty | unmeasured by focused test | n/a | n/a |
| B3 | record_stale | unmeasured by focused test | n/a | n/a |
| B4 | insufficient_streak message | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability` | n/a | yes |
| B5 | token_observation_missing | unmeasured by focused test | n/a | n/a |
| B6 | token_refresh_missing | unmeasured by focused test | n/a | n/a |
| B7 | required_endpoint_missing | unmeasured by focused test | n/a | n/a |
| B8 | completeness_failed | unmeasured by focused test | n/a | n/a |
| B9 | live_endpoint_uncovered fixed message | `TestA063DashboardTellsAnIssuedRenewalThatLostTheSupervisedScope` | yes (`analysis/scope-absorb/red.txt`) | yes |
| B10 | default generic message not used for the new code | `TestA063DashboardTellsAnIssuedRenewalThatLostTheSupervisedScope` | n/a | yes |
