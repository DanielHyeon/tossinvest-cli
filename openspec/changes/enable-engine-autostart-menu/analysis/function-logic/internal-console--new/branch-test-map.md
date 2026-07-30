# Branch Test Map: `New`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing StartVerify | existing `New` refusal test | yes | yes |
| B2 | default/custom clock | existing clock tests | yes | yes |
| B3 | invalid/valid remote contract | remote constructor tests | yes | yes |
| B4 | nil/non-nil output | existing constructor tests | yes | yes |
| B5 | nil/custom binary fingerprint | system-update tests | yes | yes |
| B6 | empty initial note | existing engine page tests | no | no |
| B7 | non-empty initial note is displayed | `TestInitialEngineNoteIsRendered` | no | no |
