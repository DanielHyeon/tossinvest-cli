# Branch Test Map: `ThresholdRegistry.LoadThresholdSet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | evidence/activation binding invalid | binding failure matrix | registry API absent | zero + error |
| B2 | same version reused with different canonical digest | `TestThresholdRegistryRejectsSameVersionWithDifferentCanonicalDigest` | registry API absent | zero + same-version error |
