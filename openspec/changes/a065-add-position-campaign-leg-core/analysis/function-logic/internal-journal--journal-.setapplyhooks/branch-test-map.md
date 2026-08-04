# Branch Test Map: `(*Journal).SetApplyHooks`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no apply function is supplied | `TestApplyHooksAreBoundOnce` | existing coverage | PASS, full journal suite |
| B1-extension | Campaign participates in a valid supplied hook set | `TestCampaignHookRunsBetweenProjectionAndExitAndRollsBackAtomically` | Campaign field absent | PASS |
| B2 | any Project/Campaign/Exit hook set is already bound | `TestApplyHooksAreBoundOnce` | existing coverage | PASS, full journal suite |
| end | Project+Campaign+Exit bind once | `TestCampaignHookRunsBetweenProjectionAndExitAndRollsBackAtomically` | Campaign field absent | PASS |
