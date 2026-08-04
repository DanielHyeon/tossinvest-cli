# Branch Test Map: `insertExactStrategyDecision`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ID-only, digest-only, malformed, whitespace or oversized reference is rejected before SQL | `TestStrategyEvidenceLineageRejectsPartialOrMalformedReference` | compile-fail contract captured | PASS |
| B2 | SQL insertion failure is returned and caller transaction cannot partially commit | existing `TestStrategyProductionIssuanceFailureRollsBackReservationAndAllAuthority` | existing coverage | PASS |
| B3 | same decision identity with changed snapshot ID or digest is a collision | `TestStrategyEvidenceLineageReplayIsExact` | compile-fail contract captured | PASS |

Exact ID/digest persistence, legacy blank-pair compatibility and exact idempotent replay are covered by
`TestStrategyEvidenceLineagePersistsOnlyImmutableReference`, the existing
`TestStrategyPlanIsAtomicExactAndIdempotent`, and `TestStrategyEvidenceLineageReplayIsExact`; all PASS.
