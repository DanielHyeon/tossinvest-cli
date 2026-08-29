# Branch Test Map: `Console.handleStart`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if r.PostFormValue("mode") == startModeRedo {` (internal/console/pages.go:168) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
| B2 | `} else if snap := c.readVerify(); snap.Present && len(snap.Pending) == 0 {` (internal/console/pages.go:180) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
| B3 | `if err != nil {` (internal/console/pages.go:170) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
| B4 | `if len(set) == 0 {` (internal/console/pages.go:174) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
| B5 | `} else if snap := c.readVerify(); snap.Present && len(snap.Pending) == 0 {` (internal/console/pages.go:180) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
| B6 | `if len(snap.Redo) > 0 {` (internal/console/pages.go:187) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
| B7 | `if _, err := c.startRun(redo); err != nil {` (internal/console/pages.go:193) | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing | yes | yes |
