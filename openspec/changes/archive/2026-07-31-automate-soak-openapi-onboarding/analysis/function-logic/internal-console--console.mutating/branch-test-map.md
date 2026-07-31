# Branch Test Map: `Console.mutating`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | GET reaches mutation | `TestTheRestartRoutesAreBehindBothGates` | existing | pending |
| B2 | remote origin differs | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` | existing | pending |
| B3 | malformed form | existing gate suite | existing | pending |
| B4 | missing/wrong CSRF | `TestTheRestartRoutesAreBehindBothGates` | existing | pending |
| B5 | bounded form exceeds limit | `TestOpenAPISetupRejectsOversizeBeforeSeams` | failed | passed |
| B6 | CSRF mismatch after bounded parse | `TestOpenAPISetupPreservesSessionAndCSRFGates` | failed | passed |
| new limit | Open API form exceeds 8 KiB | `TestOpenAPISetupRejectsOversizeBeforeSeams` | failed before implementation | pending |
