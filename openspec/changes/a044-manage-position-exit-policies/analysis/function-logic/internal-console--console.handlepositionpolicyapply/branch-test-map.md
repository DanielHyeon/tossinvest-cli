# Branch Test Map: `Console.handlePositionPolicyApply`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing engine command seam cannot mutate | console wiring suite | yes | yes |
| B2 | missing opaque engine capability is refused before command dispatch | `TestPositionPolicyReleaseCarriesEngineCapabilityAndCheckbox` | yes | yes |
| B3 | engine refusal, including missing danger confirmation, renders typed recovery | `TestPositionPolicyReleaseCarriesEngineCapabilityAndCheckbox`, `TestPositionPolicyStaleReturns412AndNeverRetries` | yes | yes |

The only browser-controlled non-capability field is the boolean `confirm=yes`; `TestPositionPolicyBrowserCarriesOnlyOpaqueEngineCapabilityAtApply` forbids mutation scope fields.
