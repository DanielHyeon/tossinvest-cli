# Branch Test Map: `BudgetCoordinator.BeginObservation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil coordinator returns zero | observation-cycle binding test | missing API | yes |
| B2 | first request initializes generation-bound endpoint state | initial observation-cycle coverage | missing API | yes |
| B3 | unavailable/exhausted/capped authority returns zero | entropy/cap focused tests | missing bounded cycle authority | yes |
| B4 | entropy read failure returns zero and latches issuance unavailable | entropy failure coverage | missing fail-closed cycle issuance | yes |
| B5 | capability collision returns zero without overwriting authority | collision coverage | missing one-shot collision guard | yes |
| success | request start captures completion watermark before later completion | `TestHeldPreCompletionResponseCannotReconcileAfterWallClockRollback` | wall timestamp reconciled held response | yes |
