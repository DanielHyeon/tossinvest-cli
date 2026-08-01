# Branch Test Map: `Repository.Insert`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | valid PLANNED insert | repository round trip | pass | pass regression |
| B2 | REGISTERING/ACTIVE insert | insert state table | accepted | pass after remediation |
| B3 | duplicate immutable identity | SQL uniqueness path | pass | pass regression |
| B4 | SQL insert fails | duplicate identity/storage error test | pass | wrapped, no partial row |
