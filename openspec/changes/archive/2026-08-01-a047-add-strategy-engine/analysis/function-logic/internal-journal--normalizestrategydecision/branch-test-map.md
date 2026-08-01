# Branch Test Map: `normalizeStrategyDecision`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | zero lineage `CreatedAt` is replaced with caller-supplied journal time after scalar trimming | production issuance success/binding tests | existing | pass |
| B2 | nonzero lineage `CreatedAt` is normalized to UTC after scalar trimming | exact plan/idempotency tests | existing | pass |
| Invariant | `DecisionPayload` is deliberately not trimmed; leading/trailing bytes remain for strict verifier refusal | `TestStrategyProductionIssuanceRejectsNonCanonicalDecisionPayload` | payload was trimmed and accepted | pass |
