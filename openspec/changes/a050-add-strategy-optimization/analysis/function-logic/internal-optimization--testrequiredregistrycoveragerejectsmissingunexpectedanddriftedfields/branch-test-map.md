# Branch Test Map: `TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact valid manifest succeeds | this test | exact coverage validator absent at `948e721` | PASS |
| B2 | empty, blank identity, missing, unexpected, wrong owner/category, and duplicate cases all execute | this test | exact coverage validator absent at `948e721` | PASS |
| B3 | each invalid case returns typed `ErrInvalidRegistry` | this test | exact coverage validator absent at `948e721` | PASS |
