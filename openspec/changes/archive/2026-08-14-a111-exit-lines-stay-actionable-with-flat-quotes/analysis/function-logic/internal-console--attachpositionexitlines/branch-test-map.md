# Branch Test Map: `attachPositionExitLines`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Every joined row is projected from the same response authority | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | prior A111 convergence RED | focused A111 suite GREEN |
| B2 | Pending desired state with no engine status remains unknown | `TestPositionsShowRuntimeUnknownWithoutDesiredFallback` | preservation coverage | existing suite GREEN |
| B3 | No stored exit yields reference-only display | `TestUSPendingAndBlockedShowEffectiveStopPlanWithoutPrice` | preservation coverage | existing suite GREEN |
| B4 | Cross-generation evidence is suppressed | `TestPositionsSuppressCrossLifecycleExitEvidence` | preservation coverage | existing suite GREEN |
| B5 | Released lifecycle closes the protection line | `TestReleasedDesignatedRowDoesNotShowPendingFallback` | preservation coverage | existing suite GREEN |
| B6 | Released legacy evidence remains nonactionable context | `TestReleasedDesignatedRowDoesNotShowPendingFallback` | preservation coverage | existing suite GREEN |
| B7 | Quarantine overrides otherwise fresh evidence | `TestAQuarantinedPositionIsNotDrawnAsProtected` | preservation coverage | existing suite GREEN |
| B8 | Quarantine without a snapshot stays explicitly unknown | `TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted` | preservation coverage | existing suite GREEN |
| B9 | Canonical snapshot is copied but stale/stopped reasons still close all values | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` | intentional A111 RED | focused A111 suite GREEN |
| B10 | Unverified lifecycle clears raw and primary line | `TestPositionsSuppressCorruptAndLifecycleUnverifiedRawEvidence` | preservation coverage | existing suite GREEN |
| B11 | Admitted legacy raw is separated from the primary line | `TestPositionsLeadWithManagedLegacyReferenceAcrossKRAndUS` | preservation coverage | existing suite GREEN |
| Rollback | `/positions` and `/dashboard` preserve a stopped marker verdict through wall rollback | `TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback` | intentional A111 RED | focused A111 suite GREEN |
