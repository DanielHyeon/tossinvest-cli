# Branch Test Map: `Console.handlePositionManagement`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | disconnected page remains read-only | console wiring tests | yes | yes |
| B2 | list error page remains read-only | console error tests | yes | yes |
| B3 | every state is rendered as one typed row | position management tests | yes | yes |
| B4 | managed states receive fixed policy actions | `TestPositionManagementOffersReleaseOnlyForExternalAdoption` | yes | yes |
| B5 | released branch is evaluated separately | readopt eligibility tests | yes | yes |
| B6 | policy choices iterate the fixed server registry | policy choice tests | yes | yes |
| B7 | RELEASE appears only for external-adoption lifecycle eligibility | `TestPositionManagementOffersReleaseOnlyForExternalAdoption` | yes | yes |
| B8 | READOPT appears only for released external lifecycle | `TestPositionManagementDoesNotOfferReadoptForIneligibleHolding` | yes | yes |
