# Branch Test Map: `Console.handlePositionPolicyPreview`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing engine command seam cannot preview | console wiring suite | yes | yes |
| B2 | tampered server-rendered selection token is refused | `TestPositionPolicyTokenTamperAndCSRFFailBeforeCommander` | yes | yes |
| B3 | engine preview error renders typed refusal | `TestPositionPolicyActiveExitConflictIs409` | yes | yes |
| B4 | missing engine-issued capability blocks apply-form rendering | preview capability contract | yes | yes |
| B5 | only danger actions receive the three-second UI wait | `TestPositionPolicyBrowserCarriesOnlyOpaqueEngineCapabilityAtApply`, `TestPositionPolicyReleaseCarriesEngineCapabilityAndCheckbox` | yes | yes |
