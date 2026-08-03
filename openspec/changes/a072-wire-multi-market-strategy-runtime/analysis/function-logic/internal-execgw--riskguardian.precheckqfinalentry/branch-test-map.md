# Branch Test Map: `RiskGuardian.PrecheckQFinalEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Guardian/request is incomplete | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B2 | policy version or digest mismatch | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B3 | snapshot evidence is stale | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B4 | market, currency, or symbol is noncanonical | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B5 | candidate quantity is nonpositive | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B6 | owner is not the exact q_final owner | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B7 | reserve/account currency scope is invalid | existing QFinalPrecheck tests | yes (opaque authority API absent) | yes |
| B8 | missing/forged public FX DTO has no opaque capability | TestQFinalPrecheckRejectsCallerConstructedFXEvidenceWithoutOpaqueAuthority | yes (public DTO was accepted) | yes |
| B9 | opaque FX cannot bind at the Guardian clock | TestQFinalRejectsCallerConstructedFXEvidence | yes (public source label was trusted) | yes |
| B10 | strategy entry cap calculation refuses | existing q_final suite | yes (opaque authority API absent) | yes |
| B11 | candidate exposure calculation refuses | existing q_final suite | yes (opaque authority API absent) | yes |
| B12 | five-bucket admission refuses | existing q_final suite | yes (opaque authority API absent) | yes |
| B13 | Guardian chain evaluation refuses | existing q_final suite | yes (opaque authority API absent) | yes |
| B14 | entry exposure value refuses | existing q_final suite | yes (opaque authority API absent) | yes |
