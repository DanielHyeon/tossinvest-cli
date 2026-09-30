# Branch Test Map: `TestARefusalReArmsTheLevel`

- Source: `internal/journal/exit_state_test.go`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:381` `if err := j.AttachExitIntent(ctx, p.ID, "i-refused"); err != nil {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B2 | `:386` `if err := j.ResolveExitProposal(ctx, p.ID, "i-refused", ProposalRefused); err != nil {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B3 | `:390` `if state.Pending() {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B4 | `:393` `if state.TakenRatioTotal != "0" {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B5 | `:397` `if err := j.RecordExitJudgement(ctx, ExitJudgement{` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B6 | `:408` `if err != nil {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B7 | `:412` `for _, e := range events {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B8 | `:413` `if e.Action == ExitEventProposalRefused {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
| B9 | `:417` `if !refused {` | — | `TestARefusalReArmsTheLevel` | n/a | yes |
