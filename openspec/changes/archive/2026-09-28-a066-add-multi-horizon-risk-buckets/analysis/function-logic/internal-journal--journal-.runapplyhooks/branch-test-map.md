# Branch Test Map: `(*Journal).runApplyHooks`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no hooks remains a no-op | existing `RecordFill` regressions | existing | GREEN in focused apply-hook regressions |
| B2 | Project hook is configured | existing apply-hook projection regressions | existing | GREEN in focused apply-hook regressions |
| B3 | Project error rolls back fill and sidecars | existing apply-hook crash regression | existing | GREEN in focused apply-hook regressions |
| B4 | Campaign hook is configured and owner bind follows it | `TestRunApplyHooksBindsRiskBucketOwnerAfterCampaignInSameTransaction` | compile failure: missing lifecycle adapter | GREEN |
| B5 | Campaign error rolls back all projections before owner bind | existing campaign apply regressions | existing | GREEN in focused apply-hook regressions |
| B6 | owner bind storage error is wrapped; semantic missing authority is absorbed/latches in the callee and Exit continues | `TestRiskBucketFillHookLatchesBindGapWithoutReturningError`, `TestRiskBucketOwnerBindRefusalWritesNothingWithoutAuthoritativeGeneration` | compile failure: missing lifecycle adapter | GREEN |
| B7 | Exit remains configured after owner bind | `TestRiskBucketFillHookLatchesBindGapWithoutReturningError` plus existing exit apply tests | existing | GREEN |
| B8 | Exit error remains atomic | existing exit rollback tests | existing | GREEN in focused apply-hook regressions |

Wave 2A (2026-09-25): AST re-extracted at HEAD `648df8ef` because the file changed around the function; the
function body text is identical to the 2026-08-04 revision and the old/new branch alignment is identical, so
the rows above are unchanged. Package suite `go test -count=1 ./internal/journal/...` PASS at `d72bc401` (495.5s)
and with `-coverprofile` at `648df8ef` (641.1s).
