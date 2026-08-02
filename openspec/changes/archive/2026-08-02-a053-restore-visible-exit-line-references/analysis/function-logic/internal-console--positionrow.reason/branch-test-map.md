# Branch Test Map: `positionRow.Reason`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | first-match explanation ordering | portfolio/console suite | pre-existing | yes |
| B2 | journal unknown or broker-only | portfolio notices tests | pre-existing | yes |
| B3 | US adoption pending or reconcile blocked | `TestUSPendingAndBlockedShowEffectiveStopPlanWithoutPrice` | yes | yes |
| B4 | desired-only US row versus already-managed or operator-released designated row | unavailable-commander test, `TestManagedDesignatedRowDoesNotShowPendingFallback`, and `TestReleasedDesignatedRowDoesNotShowPendingFallback` | yes | yes |
| B5 | ineligible unmanaged row | portfolio reason tests | pre-existing | yes |
| B6 | eligible row before exit state | portfolio reason tests | pre-existing | yes |
| B7 | managed row with exit state | portfolio rendering tests | pre-existing | yes |
