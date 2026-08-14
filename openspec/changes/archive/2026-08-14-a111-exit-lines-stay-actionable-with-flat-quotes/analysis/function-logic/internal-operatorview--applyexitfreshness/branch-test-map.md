# Branch Test Map: `ApplyExitFreshness`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch `internal/operatorview/exit_freshness.go:24`: a positively stopped engine closes immediately; every other liveness state applies the inclusive 30-second persisted-evidence gate | `TestA111SharedFreshnessAppliesTheExactThirtySecondBoundToEveryNonStoppedLiveness` | intentional A111 RED before production change | asserted by focused A111 suite |
