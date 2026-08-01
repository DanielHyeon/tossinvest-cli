# Branch Test Map: `verifyStrategyRiskBinding`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | RiskIntent hash/type mismatch | existing issuance binding tests | existing | existing |
| B2 | unknown, trailing, whitespace/noncanonical payload | `TestStrategyProductionIssuanceRejectsNonCanonicalDecisionPayload` | partial decoder accepted | pass |
| B3 | non-denormalized field changes while digest is recomputed | `TestStrategyProductionIssuanceBindsFullDecisionRecordIdentity` | partial decoder accepted | pass |
| B4 | every denormalized record field diverges | `TestStrategyProductionIssuanceRejectsEveryDivergentLineageProjection` | subset accepted | pass |
| B5 | exact canonical record succeeds atomically | `TestStrategyProductionIssuanceCommitsAuthorityReservationAndLineageTogether` | existing | existing |
| B6 | canonical remarshal fails or bytes differ | noncanonical payload table | partial decoder accepted | pass |
| B7 | identity canonical encoding failure is fail closed | fixed encodable scalar schema + vet | not dynamically triggerable | structural proof |
| B8 | payload digest, full identity, denormalized or RiskIntent field differs | projection/full-record mutation tests | subset accepted | pass |
