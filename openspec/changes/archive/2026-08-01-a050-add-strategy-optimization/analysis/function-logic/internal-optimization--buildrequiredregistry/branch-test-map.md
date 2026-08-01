# Branch Test Map: `BuildRequiredRegistry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty frozen manifest is rejected | `TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields/empty_manifest` | exact coverage absent at `948e721` | PASS after explicit case |
| B2 | all manifest entries are normalized and indexed | `TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields` valid case | exact coverage absent at `948e721` | PASS |
| B3 | blank key or owner is rejected | coverage test blank key/owner cases | exact coverage absent at `948e721` | PASS after explicit cases |
| B4 | duplicate required key is rejected | `.../duplicate_requirement` | exact coverage absent at `948e721` | PASS |
| B5 | underlying invalid provider registry remains rejected | `TestRegistryRejectsNilEmptyAndFailingProviders` | exact coverage absent at `948e721` | PASS |
| B6 | missing or unexpected registry field count is rejected | `.../missing`, `.../unexpected` | exact coverage absent at `948e721` | PASS |
| B7 | every required field is compared | valid and drift cases in `TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields` | exact coverage absent at `948e721` | PASS |
| B8 | missing key, wrong owner, or wrong category is rejected | `.../missing`, `.../wrong_owner`, `.../wrong_category` | exact coverage absent at `948e721` | PASS |
