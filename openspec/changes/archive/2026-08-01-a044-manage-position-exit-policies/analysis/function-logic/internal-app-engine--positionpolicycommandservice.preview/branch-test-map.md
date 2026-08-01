# Branch Test Map: `PositionPolicyCommandService.Preview`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | engine preparation refuses ineligible lifecycle scope | `TestPositionPolicyServiceRequiresExternalAdoptedProvenanceForLifecycleActions` | yes | yes |
| B2 | journal preview validates exact generation/version before a grant exists | `TestPositionPolicyCapabilityRequiresPreviewIsTimedAndOneTime` plus journal CAS suite | yes | yes |
| B3 | only a successfully issued opaque engine grant is returned | `TestPositionPolicyCapabilityRequiresPreviewIsTimedAndOneTime` | yes | yes |
