# Branch Test Map: `PositionPolicyCommandService.prepare`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-lifecycle action keeps server-normalized time and no client observation | `TestPositionPolicyCapabilityRequiresPreviewIsTimedAndOneTime` | yes | yes |
| B2 | lifecycle state read error fails closed | engine repository error contract | yes | yes |
| B3 | engine-entry/unknown/ambiguous provenance cannot RELEASE or READOPT | `TestPositionPolicyServiceRequiresExternalAdoptedProvenanceForLifecycleActions` | yes | yes |
| B4 | eligible RELEASE requires no price read | `TestPositionPolicyDangerousCapabilityRequiresServerSideConfirmation` | yes | yes |
| B5 | READOPT without an authoritative price source is refused | engine command suite | yes | yes |
| B6 | configured retrier owns the price-query attempt | engine retry suite | yes | yes |
| B7 | nil retrier performs one direct authoritative query | `TestPositionPolicyServiceDerivesReadoptObservationInsideEngine` | yes | yes |
| B8 | price-query error aborts preview | engine command suite | yes | yes |
| B9 | returned quotes are scanned without accepting another symbol | `TestPositionPolicyServiceDerivesReadoptObservationInsideEngine` | yes | yes |
| B10 | matching positive quote becomes fresh t0 | `TestPositionPolicyServiceDerivesReadoptObservationInsideEngine` | yes | yes |
| B11 | no matching positive quote is refused | engine command suite | yes | yes |
| B12 | invalid configured stop falls back to conservative 5% | engine command suite | yes | yes |
| B13 | synthetic-stop derivation failure aborts preview | exit-policy validation suite | yes | yes |
| B14 | empty desired policy resolves to the validated common policy | `TestPositionPolicyServiceDerivesReadoptObservationInsideEngine` | yes | yes |
