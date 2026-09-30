# Branch Test Map: `TestARefusedProposalIsReArmedAfterARestart`

- Source: `internal/journal/exit_state_test.go`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:797` `if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {` | — | `TestARefusedProposalIsReArmedAfterARestart` | n/a | yes |
| B2 | `:803` `if err := j.AttachExitIntent(ctx, positionID, "i-restart"); err != nil {` | — | `TestARefusedProposalIsReArmedAfterARestart` | n/a | yes |
| B3 | `:806` `if err := j.ResolveExitProposal(ctx, positionID, "i-restart", ProposalRefused); err != nil {` | — | `TestARefusedProposalIsReArmedAfterARestart` | n/a | yes |
| B4 | `:812` `if err := restarted.RecordExitJudgement(ctx, ExitJudgement{` | — | `TestARefusedProposalIsReArmedAfterARestart` | n/a | yes |
