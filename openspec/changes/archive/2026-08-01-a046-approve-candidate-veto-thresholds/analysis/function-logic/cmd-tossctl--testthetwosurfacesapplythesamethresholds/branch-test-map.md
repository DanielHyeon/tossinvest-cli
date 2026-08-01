# Branch Test Map: `TestTheTwoSurfacesApplyTheSameThresholds`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | scan and constructor must use the same value | `TestTheTwoSurfacesApplyTheSameThresholds` equality assertion | existing test was green before policy change | pass |
| B2 | legacy near_high 2.0 must not be active without evidence approval | same test dormant assertion | failed: `NearHighDistancePct:2.0` | pass: all fields absent |
