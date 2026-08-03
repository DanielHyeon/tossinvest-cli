# Branch Test Map: `RiskGuardian.IssuePrecheckedQFinalEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | incomplete precheck refuses | existing q_final tests | yes (opaque authority API absent) | yes |
| B2 | policy version changed after precheck refuses | existing q_final tests | yes (opaque authority API absent) | yes |
| B3 | final-clock FX authority expiry/tamper refuses before journal mutation | TestQFinalFinalIssuanceRechecksExpiredEvidenceAtGuardianClock | yes (final revalidation absent) | yes |
| B4 | position recollection fails closed | existing recollection tests | yes (opaque authority API absent) | yes |
| B5 | reservation usage recollection fails closed | existing recollection tests | yes (opaque authority API absent) | yes |
| B6 | atomic decision/reservation write fails closed | existing journal tests | yes (opaque authority API absent) | yes |
| B7 | journal result is incomplete | existing atomic issuance tests | yes (opaque authority API absent) | yes |
