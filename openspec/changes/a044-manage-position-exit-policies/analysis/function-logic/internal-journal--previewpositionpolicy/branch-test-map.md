# Branch Test Map: `previewPositionPolicy`

RED/GREEN evidence is the focused journal policy suite run on 2026-08-01.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | generation/version mismatch | `TestPositionPolicyOverrideCASAndAuditAreAtomic` | yes | yes |
| B2 | pending/CLOSING scope is recognized | `TestPositionPolicyReleaseRejectsPendingExitWithoutMutation` | yes | yes |
| B3 | lifecycle action in active exit is rejected | `TestPositionPolicyReleaseRejectsPendingExitWithoutMutation` | yes | yes |
| B4 | position-state switch is exhaustive | `TestPositionPolicyRejectsFreeActorReasonAndUnknownState` | yes | yes |
| B5 | operational position states proceed | position policy suite | yes | yes |
| B6 | unknown state fails closed | `TestPositionPolicyRejectsFreeActorReasonAndUnknownState` | yes | yes |
| B7 | virtual version gets an observation boundary | `TestPositionPolicyOverrideCASAndAuditAreAtomic` | yes | yes |
| B8 | action switch is exhaustive | position policy suite | yes | yes |
| B9 | OVERRIDE branch | `TestPositionPolicyOverrideCASAndAuditAreAtomic` | yes | yes |
| B10 | OVERRIDE requires MANAGED | `TestPositionPolicyRejectsFreeActorReasonAndUnknownState` | yes | yes |
| B11 | INHERIT branch | position policy suite | yes | yes |
| B12 | INHERIT requires MANAGED | `TestPositionPolicyRejectsFreeActorReasonAndUnknownState` | yes | yes |
| B13 | RELEASE branch | `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration` | yes | yes |
| B14 | RELEASE requires external-adoption provenance | `TestPositionPolicyReleaseAndReadoptRequireExternalAdoptionAtJournalBoundary` | yes | yes |
| B15 | RELEASE requires MANAGED | position policy suite | yes | yes |
| B16 | READOPT branch resets effective policy to the actual reset policy | `TestPositionPolicyReadoptResetsExitStateAtFreshT0` | yes | yes |
| B17 | READOPT requires external-adoption provenance | `TestPositionPolicyReleaseAndReadoptRequireExternalAdoptionAtJournalBoundary` | yes | yes |
| B18 | READOPT requires RELEASED | position policy suite | yes | yes |
| B19 | unknown action fails closed | `TestPositionPolicyRejectsFreeActorReasonAndUnknownState` | yes | yes |
| B20 | empty READOPT policy resolves to the actual ratchet reset identity | `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration` | yes | yes |
