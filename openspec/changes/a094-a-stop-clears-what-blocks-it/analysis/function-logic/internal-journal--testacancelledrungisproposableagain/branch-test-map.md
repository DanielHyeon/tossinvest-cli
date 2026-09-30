# Branch Test Map: `TestACancelledRungIsProposableAgain`

- Source: `internal/journal/exit_state_test.go`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:439` `if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
| B2 | `:443` `if _, err := j.OpenExitState(ctx, ExitStateSeed{` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
| B3 | `:450` `if err := j.RecordExitJudgement(ctx, ExitJudgement{` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
| B4 | `:457` `if got := exitStateOf(t, j, p.ID); got.ActiveRung != 1 {` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
| B5 | `:461` `if err := j.ResolveExitProposal(ctx, p.ID, "i-rung", ProposalCancelled); err != nil {` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
| B6 | `:465` `if state.ActiveRung != 0 {` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
| B7 | `:469` `if state.Baseline != "70700" {` | — | `TestACancelledRungIsProposableAgain` | n/a | yes |
