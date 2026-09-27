# Branch Test Map: `BuildExitLine`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no snapshot renders `근거 없음` and its typed reason | `TestBuildExitLineFailsClosedForStaleAndUnknownEvidence` (existing) | n/a | pass |
| B2 | an age-stale snapshot still renders `오래된 평가`; a stopped engine renders `엔진 정지`; a quarantined position renders `판정 격리`; every actionable value stays an em dash in all three | `TestBuildExitLineFailsClosedForStaleAndUnknownEvidence` (existing); `TestAStoppedEngineClosesTheProtectionLine`; `TestAQuarantinedPositionIsNotDrawnAsProtected`; both a077 tests plus the existing stale fixture | yes | yes |
| B3 | a one-share state-only projection gets its wording | `TestBuildExitLineExplainsAOneShareStateOnlyPartial` (existing) | n/a | pass |
