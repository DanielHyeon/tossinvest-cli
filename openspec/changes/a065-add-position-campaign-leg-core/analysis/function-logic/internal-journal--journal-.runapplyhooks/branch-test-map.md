# Branch Test Map: `(*Journal).runApplyHooks`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no hooks bound | existing fill no-hook suite | existing coverage | PASS, full journal suite |
| B2 | Project hook present | apply hook suite | existing | PASS |
| B3 | Project hook returns error | `TestAFailingApplyHookRollsBackTheFill` | existing | PASS |
| B4 | Campaign hook present | campaign hook atomicity test | yes | PASS |
| B5 | Campaign hook returns error | campaign hook rollback test | yes | PASS |
| B6 | Exit hook present | exit hook suite | existing | PASS |
| B7 | Exit hook returns error | `TestAFailingExitHookRollsBackTheProjectionToo` | existing | PASS |
| B2/B3 | Position succeeds/fails | `TestAFailingApplyHookRollsBackTheFill` | existing coverage | PASS |
| Campaign extension | Campaign succeeds/fails after Position | `TestCampaignHookRunsBetweenProjectionAndExitAndRollsBackAtomically` | Campaign branch absent | PASS; Position and fill rollback together |
| B4/B5 | Exit succeeds/fails after Position/Campaign | `TestAFailingExitHookRollsBackTheProjectionToo` | existing coverage | PASS |
| late fill | predecessor positive delta updates Position+watermark+replacement atomically | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | no campaign watermark hook | PASS; retry delta zero |
| terminal delta-zero | zero/partial unchanged cumulative terminal observation | `TestCampaignZeroAndPartialUnchangedTerminalObservationsCancelResidualIdempotently` | terminal observation skipped | PASS; CANCELLED and duplicate no-op |
| successor query error | lookup failure must not mean “no successor” | `TestHasCampaignSuccessorPropagatesQueryFailure` | error collapsed to false | PASS; typed error propagates |
