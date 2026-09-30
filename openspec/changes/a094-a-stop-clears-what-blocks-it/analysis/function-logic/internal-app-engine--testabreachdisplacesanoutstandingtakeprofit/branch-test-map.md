# Branch Test Map: `TestABreachDisplacesAnOutstandingTakeProfit`

- Source: `internal/app/engine/exitloop_test.go`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:850` `if cycle.Err != nil {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B2 | `:853` `if len(h.submit.cancels) == 0 {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B3 | `:858` `if len(h.submit.places) != 1 {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B4 | `:861` `if cycle := h.observe(); cycle.Err != nil {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B5 | `:864` `if len(h.submit.places) != 2 {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B6 | `:868` `if state.PendingAction != string(exitpolicy.ActionBaselineBreach) {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B7 | `:871` `if state.PendingIntentID != h.submit.places[1].IntentID {` | — | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
