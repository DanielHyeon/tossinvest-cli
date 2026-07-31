# Branch Test Map: `evidenceKey`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing id/account | orders missing-scope tests | RED scoped API compile failure | yes |
| B2 | malformed timestamp | orders timestamp tests | existing | yes |
| B3 | unsupported market | market fallback tests | existing | yes |
| B4 | market day conversion | internal/clock tests | existing | yes |
