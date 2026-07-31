# Branch Test Map: `candidateVetoThresholds`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no human evidence activation exists, so scan and `/signals` share an all-absent threshold set and `passed=0`; legacy 2.0 remains descriptor provenance only | `cmd/tossctl.TestTheTwoSurfacesApplyTheSameThresholds`; a046 descriptor/render tests | failed with runtime `near_high=2.0` | pass with all three runtime fields absent |
