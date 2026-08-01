# Branch Test Map: `TestAddResetDeltaRejectsDurationConversionOverflow`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | conversion above duration capacity fails closed | same function | unsafe multiplication could wrap | yes |
