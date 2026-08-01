# Branch Test Map: `CoreRegistry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path constructs the exact approved a041 manifest; coverage validation refuses drift | `TestCoreRegistryCoversExactApprovedSettingmetaManifest`, `TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields` | coverage validator absent at `948e721` | PASS |
