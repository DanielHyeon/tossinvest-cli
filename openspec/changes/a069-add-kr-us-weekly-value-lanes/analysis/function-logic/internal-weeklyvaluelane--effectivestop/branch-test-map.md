# Branch Test Map: `effectiveStop`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | stale or tampered candidate rejected | TestEvaluationRejectsStaleOrUnsealedStop | yes | yes |
| B2 | saved stop never retreats | TestStopCapInvalidationAndCommonExitAuthorityRemainSeparated | existing | existing |
| B3 | tighter candidate accepted | TestEffectiveStopUsesFreshTighterCandidate | yes | yes |
| B4 | invalid saved stop rejected | TestStopCapInvalidationAndCommonExitAuthorityRemainSeparated | existing | existing |
| B5 | saved stop dominates weaker candidate | TestStopCapInvalidationAndCommonExitAuthorityRemainSeparated | existing | existing |
