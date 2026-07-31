# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | remote token mode registers login/logout | existing remote route suite | existing | pending |
| B2 | remote mode wraps security | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` | existing | pending |
| default | local handler returned | existing console handler suite | existing | pending |
| new GET | setup route is session protected | `TestOpenAPISetupRequiresSession` | pending | pending |
| new POST | body limiter precedes mutation parsing | `TestOpenAPISetupRejectsOversizeBeforeSeams` | pending | pending |
| new POST | origin and CSRF still gate the route | `TestOpenAPISetupPreservesMutationGates` | pending | pending |
| new HTTPS boundary | plaintext setup submission is refused before the seam | `TestOpenAPISetupRejectsPlaintextEvenWithSessionAndCSRF` | yes | pending |
