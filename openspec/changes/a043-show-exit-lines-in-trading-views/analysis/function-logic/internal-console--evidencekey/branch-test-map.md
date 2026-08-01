# Branch Test Map: `evidenceKey`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero-length id/account is invalid while whitespace-only id is opaque and valid | `TestOpaqueBrokerOrderIDsKeepDistinctOriginAndExitEvidence`; journal whitespace-only scope assertion | trimmed IDs collided | yes |
| B2 | malformed timestamp | orders timestamp tests | existing | yes |
| B3 | unsupported market | market fallback tests | existing | yes |
| B4 | market day conversion | internal/clock tests | existing | yes |
