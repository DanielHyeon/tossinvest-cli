# Branch Test Map: `evaluate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | caller bool cannot activate dormant lane | TestEvaluationRequiresSealedAuthorization | yes | yes |
| B2 | literal/tampered evidence rejected | TestDecodedEvidenceSealRejectsLiteralAndMutation | yes | yes |
| B3 | KR/US/source isolation | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B4 | stale cap and non-exact reserved quantity | TestEvaluationRejectsStaleOrQuantityMismatchedCap | yes | yes |
| B5 | stale/unsealed stop | TestEvaluationRejectsStaleOrUnsealedStop | yes | yes |
| B6 | valid KR and US independently decide | TestAuthorizedKRAndUSEvaluateIndependently | yes | yes |
| B7 | candidate identity bound | TestEvaluationRequiresSealedAuthorization | yes | yes |
| B8 | immutable plan mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B9 | evidence market mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B10 | evidence source mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B11 | structural invalidation | TestStopCapInvalidationAndCommonExitAuthorityRemainSeparated | existing | existing |
| B12 | evidence validation refusal | TestDecodedEvidenceSealRejectsLiteralAndMutation | yes | yes |
| B13 | symbol/config mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B14 | market-week market mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B15 | calendar invalid | TestReservationRejectsStaleCalendarAtCommandEvaluation | yes | yes |
| B16 | terminal/invalid leg | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B17 | invalid reservation state | TestPositiveFillCannotBypassAtomicRiskTransition | yes | yes |
| B18 | reservation missing/mismatched | TestApplyPositiveFillAtomicCommitsReservationAndRiskTogether | yes | yes |
| B19 | reservation terminal | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B20 | seven-leg exhausted | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B21 | planned quantity exhausted | TestImmutableSevenLegAllocationAndNoUpwardReallocation | existing | existing |
| B22 | cap stale/exact quantity mismatch | TestEvaluationRejectsStaleOrQuantityMismatchedCap | yes | yes |
| B23 | stop invalid/structural cap | TestEvaluationRejectsStaleOrUnsealedStop | yes | yes |
| B24 | RR refusal | TestExactCappedTargetRRInclusiveBoundaryAndStructuralStopCap | existing | existing |
| B25 | risk refusal or decision | TestPlanBoundPrivateCapFrozenFXAndCheckedRisk | existing | existing |
